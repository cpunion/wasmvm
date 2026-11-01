wit_bindgen::generate!({ path: "../wit", world: "rust-middleware", generate_all });

struct Middleware;
impl exports::wasi::http::handler::Guest for Middleware {
    async fn handle(request: wasi::http::types::Request) -> Result<wasi::http::types::Response, wasi::http::types::ErrorCode> {
        wasi::http::handler::handle(request).await
    }
}
export!(Middleware);
