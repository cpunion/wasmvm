use std::{env, fs, time::Instant};
use wasmtime::{
    Engine, Store,
    component::{Component, Linker},
};
use wasmtime_wasi::{
    FsPerms, ResourceTable, WasiCtx, WasiCtxBuilder, WasiCtxView, WasiView,
    p2::pipe::MemoryOutputPipe,
};

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
    if args.len() != 5 {
        return Err("usage: command COMPONENT PREOPEN COUNT c|rust|full".into());
    }
    let count: usize = args[3].parse()?;
    if count == 0 || count > 10000 {
        return Err("count outside 1..10000".into());
    }
    let start = Instant::now();
    let bytes = fs::read(&args[1])?;
    let load = start.elapsed().as_nanos();
    let start = Instant::now();
    let engine = Engine::default();
    let engine_ns = start.elapsed().as_nanos();
    let start = Instant::now();
    let module = Component::new(&engine, bytes)?;
    let compile = start.elapsed().as_nanos();
    let start = Instant::now();
    let output = MemoryOutputPipe::new(1 << 20);
    let mut builder = WasiCtxBuilder::new();
    builder.args(&["wasmvm", "world"]).stdout(output.clone());
    builder.preopened_dir(&args[2], "/data", FsPerms::ReadOnly)?;
    let mut store = Store::new(
        &engine,
        State {
            ctx: builder.build(),
            table: ResourceTable::new(),
        },
    );
    let mut linker = Linker::new(&engine);
    wasmtime_wasi::p2::add_to_linker_sync(&mut linker)?;
    let setup = start.elapsed().as_nanos();
    let start = Instant::now();
    let instance = linker.instantiate(&mut store, &module)?;
    let instantiate = start.elapsed().as_nanos();
    let start = Instant::now();
    let (_, parent) = instance
        .get_export(&mut store, None, "wasi:cli/run@0.2.12")
        .ok_or("missing CLI run")?;
    let (_, index) = instance
        .get_export(&mut store, Some(&parent), "run")
        .ok_or("missing run")?;
    let run = instance
        .get_func(&mut store, &index)
        .ok_or("missing function")?
        .typed::<(), (Result<(), ()>,)>(&store)?;
    let lookup = start.elapsed().as_nanos();
    let start = Instant::now();
    if run.call(&mut store, ())?.0.is_err() {
        return Err("first command failed".into());
    }
    let first = start.elapsed().as_nanos();
    for _ in 0..10 {
        if run.call(&mut store, ())?.0.is_err() {
            return Err("warmup failed".into());
        }
    }
    let start = Instant::now();
    for _ in 0..count {
        if run.call(&mut store, ())?.0.is_err() {
            return Err("command failed".into());
        }
    }
    let warm = start.elapsed().as_nanos();
    let expected = match args[4].as_str() {
        "c" => "c-command:ok\n",
        "rust" => "rust-command:ok\n",
        "full" => "",
        _ => return Err("unknown output mode".into()),
    };
    if output.contents().as_ref() != expected.repeat(count + 11).as_bytes() {
        return Err("guest output mismatch".into());
    }
    let start = Instant::now();
    drop(store);
    drop(linker);
    drop(module);
    drop(engine);
    let close = start.elapsed().as_nanos();
    println!(
        "{{\"runtime\":\"wasmtime-48.0.1-actual-wasip2-command\",\"load_ns\":{load},\"engine_ns\":{engine_ns},\"compile_ns\":{compile},\"setup_ns\":{setup},\"instantiate_ns\":{instantiate},\"lookup_ns\":{lookup},\"first_ns\":{first},\"warm_command_ns\":{},\"close_ns\":{close},\"commands\":{count}}}",
        warm as f64 / count as f64
    );
    Ok(())
}
