use std::{env, fs, time::Instant};
use wasmtime::{
    Config, Engine, Store,
    component::{Component, Linker},
};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<_> = env::args().collect();
    if args.len() != 5 && args.len() != 6 {
        return Err("usage: async_lift COMPONENT CALLS BATCH ENTRY [GUEST_CALLBACKS]".into());
    }
    let count: u32 = args[2].parse()?;
    let batch: u32 = args[3].parse()?;
    if count == 0 || batch == 0 || count > 100000 || batch > 100000 {
        return Err("count outside bounds".into());
    }
    let (expected, callbacks) = match args[4].as_str() {
        "run" => (42, batch + 1),
        "stackful" => (batch + 42, 1),
        "direct" => (batch + 42, 0),
        "churn" => (batch, 0),
        _ => return Err("unknown entry".into()),
    };
    let callbacks: u32 = if args.len() == 6 {
        args[5].parse()?
    } else {
        callbacks
    };
    let start = Instant::now();
    let bytes = fs::read(&args[1])?;
    let load = start.elapsed().as_nanos();
    let start = Instant::now();
    let runtime = tokio::runtime::Builder::new_current_thread().build()?;
    let executor = start.elapsed().as_nanos();
    let start = Instant::now();
    let mut config = Config::new();
    config
        .wasm_component_model_async(true)
        .wasm_component_model_async_stackful(true)
        .wasm_component_model_more_async_builtins(true)
        .wasm_component_model_threading(true)
        .concurrency_support(true);
    let engine = Engine::new(&config)?;
    let engine_ns = start.elapsed().as_nanos();
    let start = Instant::now();
    let module = Component::new(&engine, bytes)?;
    let compile = start.elapsed().as_nanos();
    let start = Instant::now();
    let mut store = Store::new(&engine, ());
    let linker = Linker::new(&engine);
    let instance = runtime.block_on(linker.instantiate_async(&mut store, &module))?;
    let instantiate = start.elapsed().as_nanos();
    let start = Instant::now();
    let f = instance.get_typed_func::<(u32,), (u32,)>(&mut store, &args[4])?;
    let lookup = start.elapsed().as_nanos();
    let (first, warm) = runtime.block_on(async {
        let start = Instant::now();
        if f.call_async(&mut store, (batch,)).await?.0 != expected {
            return Err("guest result mismatch".into());
        }
        let first = start.elapsed().as_nanos();
        for _ in 0..10 {
            if f.call_async(&mut store, (batch,)).await?.0 != expected {
                return Err("guest result mismatch".into());
            }
        }
        let start = Instant::now();
        for _ in 0..count {
            if f.call_async(&mut store, (batch,)).await?.0 != expected {
                return Err("guest result mismatch".into());
            }
        }
        Ok::<_, Box<dyn std::error::Error>>((
            first,
            start.elapsed().as_nanos() as f64 / count as f64,
        ))
    })?;
    let start = Instant::now();
    drop(store);
    drop(linker);
    drop(module);
    drop(engine);
    drop(runtime);
    let close = start.elapsed().as_nanos();
    println!(
        "{{\"load_ns\":{load},\"executor_ns\":{executor},\"engine_ns\":{engine_ns},\"compile_ns\":{compile},\"store_instantiate_ns\":{instantiate},\"lookup_ns\":{lookup},\"first_ns\":{first},\"warm_call_ns\":{warm},\"callbacks_per_call\":{callbacks},\"calls\":{count},\"close_ns\":{close}}}"
    );
    Ok(())
}
