use std::{env, fs, time::Instant};
use wasmtime::{
    Config, Engine, Store,
    component::{Component, Linker},
};
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<_> = env::args().collect();
    if args.len() != 4 {
        return Err("usage: error_context COMPONENT CALLS BATCH".into());
    }
    let count: u32 = args[2].parse()?;
    let batch: u32 = args[3].parse()?;
    if count == 0
        || batch == 0
        || count > 100000
        || batch > 100000
        || (u64::from(count) + 11) * u64::from(batch) >= 1 << 28
    {
        return Err("count outside bounds".into());
    }
    let start = Instant::now();
    let bytes = fs::read(&args[1])?;
    let load = start.elapsed().as_nanos();
    let start = Instant::now();
    let mut config = Config::new();
    config.wasm_component_model_error_context(true);
    let engine = Engine::new(&config)?;
    let engine_ns = start.elapsed().as_nanos();
    let start = Instant::now();
    let module = Component::new(&engine, bytes)?;
    let compile = start.elapsed().as_nanos();
    let start = Instant::now();
    let mut store = Store::new(&engine, ());
    let linker = Linker::new(&engine);
    let instance = linker.instantiate(&mut store, &module)?;
    let instantiate = start.elapsed().as_nanos();
    let start = Instant::now();
    let f = instance.get_typed_func::<(u32,), (u32,)>(&mut store, "churn")?;
    let lookup = start.elapsed().as_nanos();
    let start = Instant::now();
    if f.call(&mut store, (batch,))?.0 != batch {
        return Err("first result".into());
    }
    let first = start.elapsed().as_nanos();
    for _ in 0..10 {
        if f.call(&mut store, (batch,))?.0 != batch {
            return Err("warmup result".into());
        }
    }
    let start = Instant::now();
    for _ in 0..count {
        if f.call(&mut store, (batch,))?.0 != batch {
            return Err("warm result".into());
        }
    }
    let warm = start.elapsed().as_nanos() as f64 / (count as f64 * batch as f64);
    let start = Instant::now();
    drop(store);
    drop(module);
    drop(linker);
    drop(engine);
    let close = start.elapsed().as_nanos();
    println!(
        "{{\"load_ns\":{load},\"engine_ns\":{engine_ns},\"compile_ns\":{compile},\"store_instantiate_ns\":{instantiate},\"lookup_ns\":{lookup},\"first_ns\":{first},\"warm_context_ns\":{warm},\"close_ns\":{close}}}"
    );
    Ok(())
}
