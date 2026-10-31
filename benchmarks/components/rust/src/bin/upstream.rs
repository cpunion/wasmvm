use std::{env, fs, time::Instant};
use wasmtime::{
    Engine, Store,
    component::{Component, Linker},
};
use wasmtime_wasi::{
    FsPerms, I32Exit, ResourceTable, WasiCtx, WasiCtxBuilder, WasiCtxView, WasiView,
    p2::pipe::{MemoryInputPipe, MemoryOutputPipe},
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
fn hex(bytes: &[u8]) -> String {
    bytes.iter().map(|b| format!("{b:02x}")).collect()
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<_> = env::args().collect();
    if args.len() < 3 {
        return Err("usage: upstream COMPONENT PREOPEN [GUEST_ARGS...]".into());
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
    let out = MemoryOutputPipe::new(1 << 20);
    let stderr = MemoryOutputPipe::new(1 << 20);
    let mut builder = WasiCtxBuilder::new();
    builder
        .args(&args[3..])
        .env("frabjous", "day")
        .env("callooh", "callay")
        .initial_cwd("/")
        .stdin(MemoryInputPipe::new(
            b"So rested he by the Tumtum tree".as_slice(),
        ))
        .stdout(out.clone())
        .stderr(stderr.clone());
    builder.preopened_dir(&args[2], "/", FsPerms::ReadWrite)?;
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
        .get_export(&mut store, None, "wasi:cli/run@0.2.0")
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
    let result = run.call(&mut store, ());
    let first = start.elapsed().as_nanos();
    let (code, trapped) = match result {
        Ok((Ok(()),)) => (0, false),
        Ok((Err(()),)) => (1, false),
        Err(e) => match e.downcast_ref::<I32Exit>() {
            Some(exit) => (exit.0 as u32, false),
            None => {
                eprintln!("{e:#}");
                (0, true)
            }
        },
    };
    let stdout = hex(&out.contents());
    let stderr = hex(&stderr.contents());
    let start = Instant::now();
    drop(store);
    drop(module);
    drop(linker);
    drop(engine);
    let close = start.elapsed().as_nanos();
    println!(
        "{{\"load_ns\":{load},\"engine_ns\":{engine_ns},\"compile_ns\":{compile},\"store_provider_ns\":{setup},\"instantiate_ns\":{instantiate},\"lookup_ns\":{lookup},\"first_ns\":{first},\"close_ns\":{close},\"code\":{code},\"trapped\":{trapped},\"stdout_hex\":\"{stdout}\",\"stderr_hex\":\"{stderr}\"}}"
    );
    Ok(())
}
