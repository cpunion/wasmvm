//! Optional benchmark-only C bridge to the pinned P3 provider. No owned VM code.

use bytes::Bytes;
use http_body_util::{BodyExt, Full, Limited};
use std::{
    ffi::{CString, c_char},
    future::Future,
    pin::Pin,
    ptr, slice,
    sync::Mutex,
    time::Instant,
};
use wasmtime::{
    Config, Engine, Store,
    component::{Component, Linker, ResourceTable},
};
use wasmtime_wasi::{WasiCtx, WasiCtxBuilder, WasiCtxView, WasiView};
use wasmtime_wasi_http::{
    Error, RequestOptions, WasiBody, WasiHttpCtx, WasiHttpCtxView, WasiHttpHooks, WasiHttpView,
    p3::{self, bindings::Service},
};

struct State {
    wasi: WasiCtx,
    http: WasiHttpCtx,
    table: ResourceTable,
    hooks: NoNetwork,
}
struct NoNetwork;
type Completion = Box<dyn Future<Output = Result<(), Error>> + Send>;
impl WasiHttpHooks for NoNetwork {
    fn send_request(
        &mut self,
        _: hyper::Request<WasiBody>,
        _: Option<RequestOptions>,
        _: Completion,
    ) -> Box<dyn Future<Output = Result<(hyper::Response<WasiBody>, Completion), Error>> + Send>
    {
        Box::new(async { Err(Error::HttpRequestDenied) })
    }
}
impl WasiView for State {
    fn ctx(&mut self) -> WasiCtxView<'_> {
        WasiCtxView {
            ctx: &mut self.wasi,
            table: &mut self.table,
        }
    }
}
impl WasiHttpView for State {
    fn http(&mut self) -> WasiHttpCtxView<'_> {
        WasiHttpCtxView {
            ctx: &mut self.http,
            table: &mut self.table,
            hooks: &mut self.hooks,
        }
    }
}
struct Worker {
    store: Store<State>,
    service: Service,
    runtime: tokio::runtime::Runtime,
    payload: Bytes,
    response: Bytes,
    error: CString,
    failed: bool,
}
struct Ready {
    // Drop stores/runtimes before their shared compiled module and engine.
    workers: Vec<Mutex<Worker>>,
    _module: Component,
    _engine: Engine,
}
pub struct Owner {
    ready: Option<Ready>,
    phases: [u64; 6],
    error: CString,
}

#[repr(C)]
pub struct Reply {
    bytes: *const u8,
    length: usize,
    status: u32,
    guest: u32,
    trailer: u32,
}
fn message(text: impl AsRef<str>) -> CString {
    CString::new(text.as_ref().replace('\0', " ")).unwrap()
}
fn ns(start: Instant) -> u64 {
    start.elapsed().as_nanos().try_into().unwrap()
}

fn prepare(
    bytes: &[u8],
    payload: &[u8],
    count: u32,
    phases: &mut [u64; 6],
) -> wasmtime::Result<Ready> {
    let start = Instant::now();
    let mut config = Config::new();
    config
        .cranelift_opt_level(wasmtime::OptLevel::Speed)
        .wasm_component_model_async(true)
        .wasm_component_model_async_stackful(true)
        .wasm_component_model_more_async_builtins(true)
        .wasm_component_model_threading(true)
        .concurrency_support(true);
    let engine = Engine::new(&config)?;
    phases[0] = ns(start);
    let start = Instant::now();
    let module = Component::new(&engine, bytes)?;
    phases[1] = ns(start);
    let payload = Bytes::copy_from_slice(payload);
    let mut workers = Vec::new();
    for _ in 0..count {
        let start = Instant::now();
        let runtime = tokio::runtime::Builder::new_current_thread()
            .enable_all()
            .build()?;
        phases[5] += ns(start);
        let start = Instant::now();
        let mut store = Store::new(
            &engine,
            State {
                wasi: WasiCtxBuilder::new().build(),
                http: WasiHttpCtx::new(),
                table: ResourceTable::new(),
                hooks: NoNetwork,
            },
        );
        let mut linker = Linker::new(&engine);
        wasmtime_wasi::p3::add_to_linker(&mut linker)?;
        p3::add_to_linker(&mut linker)?;
        phases[2] += ns(start);
        let start = Instant::now();
        let instance = runtime.block_on(linker.instantiate_async(&mut store, &module))?;
        phases[3] += ns(start);
        let start = Instant::now();
        let service = Service::new(&mut store, &instance)?;
        phases[4] += ns(start);
        workers.push(Mutex::new(Worker {
            store,
            service,
            runtime,
            payload: payload.clone(),
            response: Bytes::new(),
            error: message(""),
            failed: false,
        }));
    }
    Ok(Ready {
        workers,
        _module: module,
        _engine: engine,
    })
}

/// Input buffers are borrowed only for this call; returned owners require close.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn wasmvm_p3_open(
    bytes: *const u8,
    size: usize,
    payload: *const u8,
    length: usize,
    count: u32,
) -> *mut Owner {
    let mut owner = Owner {
        ready: None,
        phases: [0; 6],
        error: message(""),
    };
    if bytes.is_null()
        || payload.is_null()
        || size == 0
        || length == 0
        || length > 1 << 20
        || count == 0
        || count > 64
    {
        owner.error = message("invalid P3 reference input");
    } else {
        let result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
            prepare(
                unsafe { slice::from_raw_parts(bytes, size) },
                unsafe { slice::from_raw_parts(payload, length) },
                count,
                &mut owner.phases,
            )
        }));
        match result {
            Ok(Ok(ready)) => owner.ready = Some(ready),
            Ok(Err(error)) => owner.error = message(format!("{error:#}")),
            Err(_) => owner.error = message("P3 reference setup panic"),
        }
    }
    Box::into_raw(Box::new(owner))
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn wasmvm_p3_phase(owner: *const Owner, phase: u32) -> u64 {
    unsafe { (*owner).phases.get(phase as usize).copied().unwrap_or(0) }
}
#[unsafe(no_mangle)]
pub unsafe extern "C" fn wasmvm_p3_error(owner: *const Owner, index: u32) -> *const c_char {
    let o = unsafe { &*owner };
    if let Some(worker) = o.ready.as_ref().and_then(|r| r.workers.get(index as usize)) {
        worker
            .lock()
            .unwrap_or_else(|e| e.into_inner())
            .error
            .as_ptr()
    } else {
        o.error.as_ptr()
    }
}

fn invoke(worker: &mut Worker) -> wasmtime::Result<(Bytes, u32, bool, bool)> {
    let Worker {
        store,
        service,
        runtime,
        payload,
        ..
    } = worker;
    runtime.block_on(store.run_concurrent(async |accessor| {
        let request = hyper::Request::builder()
            .method("POST")
            .uri("http://localhost/echo")
            .body(Full::new(payload.clone()))?;
        let (request, complete) = accessor.with(|mut store| {
            let (request, complete) =
                p3::Request::from_http(store.data_mut().http().hooks, request);
            // Erase the opaque future within the borrow scope. Rust2024's
            // capture rules otherwise retain the temporary Access lifetime.
            let complete: Pin<
                Box<dyn Future<Output = Result<(), wasmtime_wasi_http::Error>> + Send>,
            > = Box::pin(complete);
            (request, complete)
        });
        let response = service.handle(accessor, request).await??;
        let response = accessor.with(|mut store| response.into_http(&mut store, complete))?;
        let status = response.status().as_u16() as u32;
        let guest = response
            .headers()
            .get("x-guest")
            .is_some_and(|v| v == "rust-p3");
        let collected = Limited::new(response.into_body(), payload.len() + 1)
            .collect()
            .await
            .map_err(|error| wasmtime::Error::msg(error.to_string()))?;
        let trailer = collected
            .trailers()
            .and_then(|v| v.get("x-guest-trailer"))
            .is_some_and(|v| v == "done");
        Ok::<_, wasmtime::Error>((collected.to_bytes(), status, guest, trailer))
    }))?
}

/// Each worker is serialized. Reply bytes remain valid until its next call or close.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn wasmvm_p3_call(owner: *const Owner, index: u32, reply: *mut Reply) -> u32 {
    if reply.is_null() {
        return 1;
    }
    unsafe {
        ptr::write(
            reply,
            Reply {
                bytes: ptr::null(),
                length: 0,
                status: 0,
                guest: 0,
                trailer: 0,
            },
        );
    }
    let o = unsafe { &*owner };
    let Some(worker) = o.ready.as_ref().and_then(|r| r.workers.get(index as usize)) else {
        return 1;
    };
    let mut worker = worker.lock().unwrap_or_else(|e| e.into_inner());
    if worker.failed {
        return 1;
    }
    match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| invoke(&mut worker))) {
        Ok(Ok((body, status, guest, trailer))) => {
            worker.response = body;
            unsafe {
                ptr::write(
                    reply,
                    Reply {
                        bytes: worker.response.as_ptr(),
                        length: worker.response.len(),
                        status,
                        guest: guest.into(),
                        trailer: trailer.into(),
                    },
                );
            }
            0
        }
        Ok(Err(error)) => {
            worker.error = message(format!("{error:#}"));
            worker.failed = true;
            1
        }
        Err(_) => {
            worker.error = message("P3 reference invocation panic");
            worker.failed = true;
            1
        }
    }
}

/// Caller must join every worker before closing the owner.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn wasmvm_p3_close(owner: *mut Owner) {
    if !owner.is_null() {
        drop(unsafe { Box::from_raw(owner) });
    }
}
