use std::{env, fs, time::Instant};
use wasmtime::{
    Engine, Store,
    component::{Component, Linker, ResourceTable},
};
use wasmtime_wasi::{WasiCtx, WasiCtxBuilder, WasiCtxView, WasiView};
use wasmtime_wasi_http::{WasiHttpCtx, WasiHttpCtxView, WasiHttpView};

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
    if args.len() != 3 {
        return Err("usage: http_fields COMPONENT COUNT".into());
    }
    let count: u64 = args[2].parse()?;
    if count == 0 {
        return Err("zero count".into());
    }
    let t = Instant::now();
    let bytes = fs::read(&args[1])?;
    let load = t.elapsed().as_nanos();
    let t = Instant::now();
    let engine = Engine::default();
    let engine_time = t.elapsed().as_nanos();
    let t = Instant::now();
    let component = Component::new(&engine, bytes)?;
    let compile = t.elapsed().as_nanos();
    let t = Instant::now();
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
    let setup = t.elapsed().as_nanos();
    let t = Instant::now();
    let instance = linker.instantiate(&mut store, &component)?;
    let instantiate = t.elapsed().as_nanos();
    let probe = instance.get_typed_func::<(), (u32,)>(&mut store, "probe")?;
    if probe.call(&mut store, ())?.0 != 42 {
        return Err("field probe failed".into());
    }
    let t = Instant::now();
    let batch = instance.get_typed_func::<(u32,), (u32,)>(&mut store, "fields-batch")?;
    let lookup = t.elapsed().as_nanos();
    let t = Instant::now();
    if batch.call(&mut store, (128,))?.0 != 128 {
        return Err("first failed".into());
    };
    let first = t.elapsed().as_nanos();
    for _ in 0..10 {
        if batch.call(&mut store, (128,))?.0 != 128 {
            return Err("warmup failed".into());
        }
    }
    let mut checksum: u64 = 0;
    let t = Instant::now();
    for _ in 0..count {
        let n = batch.call(&mut store, (128,))?.0;
        if n != 128 {
            return Err("batch failed".into());
        };
        checksum += u64::from(n)
    }
    let warm = t.elapsed().as_nanos();
    let t = Instant::now();
    drop(store);
    drop(linker);
    drop(component);
    drop(engine);
    let close = t.elapsed().as_nanos();
    println!(
        "{{\"runtime\":\"wasmtime-48.0.1-typed-http-p2\",\"load_ns\":{load},\"engine_ns\":{engine_time},\"compile_ns\":{compile},\"setup_ns\":{setup},\"instantiate_ns\":{instantiate},\"lookup_ns\":{lookup},\"first_ns\":{first},\"warm_ns_per_call\":{},\"warm_fields_lifetime_ns\":{},\"lifetimes_per_call\":128,\"close_ns\":{close},\"checksum\":{checksum}}}",
        warm as f64 / count as f64,
        warm as f64 / count as f64 / 128.0
    );
    Ok(())
}
