use std::{env, fs, time::Instant};
use wasmtime::{
    Engine, Module, Store,
    component::{Component, Linker},
};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<_> = env::args().collect();
    if args.len() != 4 {
        return Err("expected component, Core module, positive iteration count".into());
    }
    let count: usize = args[3].parse()?;
    if count == 0 {
        return Err("iteration count must be positive".into());
    }
    let start = Instant::now();
    let bytes = fs::read(&args[1])?;
    let core_bytes = fs::read(&args[2])?;
    let load = start.elapsed().as_nanos();
    let start = Instant::now();
    let engine = Engine::default();
    let engine_ns = start.elapsed().as_nanos();
    let start = Instant::now();
    let module = Module::from_binary(&engine, &core_bytes)?;
    let core_compile = start.elapsed().as_nanos();
    let start = Instant::now();
    let component = Component::from_binary(&engine, &bytes)?;
    let compile = start.elapsed().as_nanos();
    let start = Instant::now();
    let mut store = Store::new(&engine, ());
    let mut linker = Linker::new(&engine);
    linker.root().module("math", &module)?;
    linker.root().func_wrap("mark", |_, (): ()| Ok(()))?;
    drop(module);
    let setup = start.elapsed().as_nanos();
    let start = Instant::now();
    let instance = linker.instantiate(&mut store, &component)?;
    let instantiate = start.elapsed().as_nanos();
    let start = Instant::now();
    let function = instance.get_typed_func::<(u32, u32), (u32,)>(&mut store, "add")?;
    let lookup = start.elapsed().as_nanos();
    let start = Instant::now();
    let (result,) = function.call(&mut store, (37, 5))?;
    let first = start.elapsed().as_nanos();
    assert_eq!(result, 42);
    let mut checksum = 0u64;
    let start = Instant::now();
    for i in 0..count {
        let (result,) = function.call(&mut store, (i as u32, 5))?;
        assert_eq!(result, (i as u32).wrapping_add(5));
        checksum += result as u64;
    }
    let warm = start.elapsed().as_nanos() as f64 / count as f64;
    println!(
        "{{\"load_ns\":{load},\"engine_ns\":{engine_ns},\"core_compile_ns\":{core_compile},\"component_compile_ns\":{compile},\"store_linker_ns\":{setup},\"instantiate_ns\":{instantiate},\"lookup_ns\":{lookup},\"first_ns\":{first},\"warm_ns_per_call\":{warm:.3},\"count\":{count},\"checksum\":{checksum}}}"
    );
    Ok(())
}
