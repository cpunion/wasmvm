use bytes::Bytes;
use http_body_util::{BodyExt, Full};
use std::{
    env, fs,
    sync::{Arc, Mutex},
};
use wasmtime::{
    Engine, Store,
    component::{Component, Linker, Resource, ResourceTable},
};
use wasmtime_wasi::{WasiCtx, WasiCtxBuilder, WasiCtxView, WasiView};
use wasmtime_wasi_http::{
    WasiHttpCtx, WasiHttpCtxView, WasiHttpView,
    p2::{
        bindings::http::types::Scheme,
        types::{HostIncomingRequest, HostResponseOutparam},
    },
};
struct State {
    wasi: WasiCtx,
    http: WasiHttpCtx,
    table: ResourceTable,
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
            hooks: Default::default(),
        }
    }
}
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<_> = env::args().collect();
    if args.len() != 2 {
        return Err("usage: http_proxy COMPONENT".into());
    }
    let engine = Engine::default();
    let component = Component::new(&engine, fs::read(&args[1])?)?;
    let mut store = Store::new(
        &engine,
        State {
            wasi: WasiCtxBuilder::new().build(),
            http: WasiHttpCtx::new(),
            table: ResourceTable::new(),
        },
    );
    let mut linker = Linker::new(&engine);
    wasmtime_wasi::p2::add_to_linker_sync(&mut linker)?;
    wasmtime_wasi_http::p2::add_only_http_to_linker_sync(&mut linker)?;
    let instance = linker.instantiate(&mut store, &component)?;
    let request = hyper::Request::builder()
        .method("POST")
        .uri("http://localhost/echo")
        .body(Full::new(Bytes::from("0123456789abcdef".repeat(256))))?;
    let incoming = store
        .data_mut()
        .http()
        .new_incoming_request(Scheme::Http, request)?;
    let slot = Arc::new(Mutex::new(None));
    let saved = slot.clone();
    let out = store
        .data_mut()
        .http()
        .new_response_outparam_from_callback(move |value| *saved.lock().unwrap() = Some(value))?;
    let (_, parent) = instance
        .get_export(&mut store, None, "wasi:http/incoming-handler@0.2.12")
        .ok_or("missing handler")?;
    let (_, index) = instance
        .get_export(&mut store, Some(&parent), "handle")
        .ok_or("missing handle")?;
    let function = instance
        .get_func(&mut store, &index)
        .ok_or("missing function")?
        .typed::<(
            Resource<HostIncomingRequest>,
            Resource<HostResponseOutparam>,
        ), ()>(&store)?;
    function.call(&mut store, (incoming, out))?;
    let response = slot
        .lock()
        .unwrap()
        .take()
        .ok_or("missing response")?
        .map_err(|e| format!("response error: {e:?}"))?;
    if response.status() != 200 {
        return Err("bad status".into());
    }
    let runtime = tokio::runtime::Runtime::new()?;
    let collected = runtime.block_on(response.into_body().collect())?;
    if collected
        .trailers()
        .and_then(|t| t.get("x-reply-trailer"))
        .and_then(|v| v.to_str().ok())
        != Some("done")
    {
        return Err("missing trailers".into());
    }
    if collected.to_bytes() != Bytes::from("0123456789abcdef".repeat(256)) {
        return Err("body mismatch".into());
    }
    println!(
        "actual Wasmtime 48.0.1 HTTPp2 incoming handler: status, 4096-byte body and trailers passed"
    );
    Ok(())
}
