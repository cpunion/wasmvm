use std::{env, fs, time::Instant};
use wasmtime::{
    Config, Engine, Store,
    component::{Component, Linker},
};
use wasmtime_wasi::{ResourceTable, WasiCtx, WasiCtxBuilder, WasiCtxView, WasiView};
struct State {
    ctx: WasiCtx,
    table: ResourceTable,
}
impl WasiView for State {
    fn ctx(&mut self) -> WasiCtxView<'_> {
        WasiCtxView {
            ctx: &mut self.ctx,
            table: &mut self.table,
        }
    }
}
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<_> = env::args().collect();
    if args.len() != 3 {
        return Err("usage: udp COMPONENT COUNT".into());
    }
    let count: u64 = args[2].parse()?;
    if count == 0 {
        return Err("zero count".into());
    }
    let t = Instant::now();
    let bytes = fs::read(&args[1])?;
    let load = t.elapsed().as_nanos();
    let t = Instant::now();
    let engine = Engine::new(&Config::new())?;
    let engine_time = t.elapsed().as_nanos();
    let t = Instant::now();
    let component = Component::new(&engine, bytes)?;
    let compile = t.elapsed().as_nanos();
    let t = Instant::now();
    let mut builder = WasiCtxBuilder::new();
    builder.inherit_network().allow_udp(true);
    let mut store = Store::new(
        &engine,
        State {
            ctx: builder.build(),
            table: ResourceTable::new(),
        },
    );
    let mut linker = Linker::new(&engine);
    wasmtime_wasi::p2::add_to_linker_sync(&mut linker)?;
    let setup = t.elapsed().as_nanos();
    let t = Instant::now();
    let instance = linker.instantiate(&mut store, &component)?;
    let instantiate = t.elapsed().as_nanos();
    let exercise = instance.get_typed_func::<(u32,), (u32,)>(&mut store, "exercise")?;
    for family in [0, 1] {
        if exercise.call(&mut store, (family,))?.0 != 42 {
            return Err("exercise failed".into());
        }
    }
    let open = instance.get_typed_func::<(u32, bool), (u32,)>(&mut store, "setup")?;
    let close = instance.get_typed_func::<(), (u32,)>(&mut store, "teardown")?;
    let t = Instant::now();
    if open.call(&mut store, (0, true))?.0 != 1 {
        return Err("setup failed".into());
    };
    let session_setup = t.elapsed().as_nanos();
    let t = Instant::now();
    let transfer = instance.get_typed_func::<(u32,), (u32,)>(&mut store, "transfer")?;
    let lookup = t.elapsed().as_nanos();
    let t = Instant::now();
    if transfer.call(&mut store, (128,))?.0 != 128 {
        return Err("first failed".into());
    };
    let first = t.elapsed().as_nanos();
    for _ in 0..10 {
        if transfer.call(&mut store, (128,))?.0 != 128 {
            return Err("warmup failed".into());
        }
    }
    let mut checksum: u64 = 0;
    let t = Instant::now();
    for _ in 0..count {
        let n = transfer.call(&mut store, (128,))?.0;
        if n != 128 {
            return Err("transfer failed".into());
        };
        checksum += u64::from(n)
    }
    let warm = t.elapsed().as_nanos();
    let t = Instant::now();
    if close.call(&mut store, ())?.0 != 1 {
        return Err("teardown failed".into());
    };
    let session_close = t.elapsed().as_nanos();
    let t = Instant::now();
    drop(store);
    drop(linker);
    drop(component);
    drop(engine);
    let teardown = t.elapsed().as_nanos();
    println!(
        "{{\"runtime\":\"wasmtime-48.0.1-typed-wasip2-udp\",\"load_ns\":{load},\"engine_ns\":{engine_time},\"compile_ns\":{compile},\"setup_ns\":{setup},\"instantiate_ns\":{instantiate},\"lookup_ns\":{lookup},\"session_setup_ns\":{session_setup},\"first_ns\":{first},\"warm_ns_per_call\":{},\"warm_roundtrip_ns\":{},\"roundtrips_per_call\":128,\"session_close_ns\":{session_close},\"close_ns\":{teardown},\"checksum\":{checksum}}}",
        warm as f64 / count as f64,
        warm as f64 / count as f64 / 128.0
    );
    Ok(())
}
