wit_bindgen::generate!({ path: "../wit", world: "rust-command", generate_all });

struct Command;
impl exports::wasi::cli::run::Guest for Command {
    fn run() -> Result<(), ()> {
        let args = wasi::cli::environment::get_arguments();
        if args.len() != 2 || args[1] != "world" {
            return Err(());
        }
        if wasmvm::world_tests::arithmetic::increment(41) != 42 {
            return Err(());
        }
        if wasi::random::random::get_random_bytes(16).len() != 16 {
            return Err(());
        }
        let poll = wasi::clocks::monotonic_clock::subscribe_duration(0);
        poll.block();
        drop(poll);
        let dirs = wasi::filesystem::preopens::get_directories();
        if dirs.len() != 1
            || dirs[0].0.stat().map_err(|_| ())?.type_
                != wasi::filesystem::types::DescriptorType::Directory
        {
            return Err(());
        }
        drop(dirs);
        let out = wasi::cli::stdout::get_stdout();
        out.blocking_write_and_flush(b"rust-command:ok\n")
            .map_err(|_| ())?;
        Ok(())
    }
}
export!(Command);
