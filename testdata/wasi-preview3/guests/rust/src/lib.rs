wit_bindgen::generate!({ path: "../wit", world: "rust-command", generate_all });

struct Command;
impl exports::wasi::cli::run::Guest for Command {
    async fn run() -> Result<(), ()> {
        let args = wasi::cli::environment::get_arguments();
        if args.len() != 2 || args[1] != "world" { return Err(()); }
        if wasi::random::random::get_random_bytes(16).len() != 16 { return Err(()); }
        wasi::clocks::monotonic_clock::wait_for(0).await;
        let dirs = wasi::filesystem::preopens::get_directories();
        if dirs.len() != 1 || !matches!(dirs[0].0.stat().await.map_err(|_| ())?.type_,
                wasi::filesystem::types::DescriptorType::Directory) { return Err(()); }
        drop(dirs);
        let (mut writer, reader) = wit_stream::new::<u8>();
        let complete = wasi::cli::stdout::write_via_stream(reader);
        if !writer.write_all(b"rust-p3-command:ok\n".to_vec()).await.is_empty() { return Err(()); }
        drop(writer);
        complete.await.map_err(|_| ())?;
        Ok(())
    }
}
export!(Command);
