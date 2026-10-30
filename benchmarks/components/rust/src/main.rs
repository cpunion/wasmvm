use std::{env, fs, time::Instant};
use wasmtime::{
    Engine, Store,
    component::{Component, Linker},
};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<_> = env::args().collect();
    if args.len() != 3 {
        return Err("expected fixture and positive iteration count".into());
    }
    let count: usize = args[2].parse()?;
    if count == 0 {
        return Err("iteration count must be positive".into());
    }
    let start = Instant::now();
    let bytes = fs::read(&args[1])?;
    let load = start.elapsed().as_nanos();
    let start = Instant::now();
    let engine = Engine::default();
    let engine_ns = start.elapsed().as_nanos();
    let start = Instant::now();
    let component = Component::from_binary(&engine, &bytes)?;
    let compile = start.elapsed().as_nanos();
    let start = Instant::now();
    let mut store = Store::new(&engine, ());
    let linker = Linker::new(&engine);
    let setup = start.elapsed().as_nanos();
    let start = Instant::now();
    let instance = linker.instantiate(&mut store, &component)?;
    let instantiate = start.elapsed().as_nanos();
    let start = Instant::now();
    let function = instance.get_typed_func::<(&[u8],), (Vec<u8>,)>(&mut store, "echo")?;
    let lookup = start.elapsed().as_nanos();
    let data: Vec<u8> = (0..16384).map(|i| i as u8).collect();
    let start = Instant::now();
    let (first_result,) = function.call(&mut store, (&data,))?;
    let first = start.elapsed().as_nanos();
    assert_eq!(first_result, data);
    drop(first_result);
    let start = Instant::now();
    for _ in 0..count {
        let (result,) = function.call(&mut store, (&data,))?;
        assert_eq!(result.len(), data.len());
        assert_eq!(result[0], 0);
        assert_eq!(result[data.len() - 1], 255);
        drop(result);
    }
    let warm = start.elapsed().as_nanos() as f64 / count as f64;
    println!(
        "{{\"load_ns\":{load},\"engine_ns\":{engine_ns},\"compile_ns\":{compile},\"store_linker_ns\":{setup},\"instantiate_ns\":{instantiate},\"lookup_ns\":{lookup},\"first_ns\":{first},\"warm_call_and_result_drop_ns\":{warm:.3},\"bytes\":{},\"count\":{count},\"checksum\":{}}}",
        data.len(),
        data.len() * count
    );
    Ok(())
}
