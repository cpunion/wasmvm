wit_bindgen::generate!({ path: "../wit", world: "rust-http", generate_all });

use wasi::http::types::{ErrorCode, Fields, Request, Response};
struct Service;
impl exports::wasi::http::handler::Guest for Service {
    async fn handle(request: Request) -> Result<Response, ErrorCode> {
        let (handled, handling) = wit_future::new(|| Ok(()));
        drop(handled);
        let (contents, trailers) = Request::consume_body(request, handling);
        let data = contents.collect().await;
        drop(trailers.await?);
        let headers = Fields::new();
        headers.append("x-guest", b"rust-p3").unwrap();
        let (mut writer, reader) = wit_stream::new::<u8>();
        let (trailer_writer, trailer_reader) = wit_future::new(|| Ok(None::<Fields>));
        let (response, transmitted) = Response::new(headers, Some(reader), trailer_reader);
        response.set_status_code(201).unwrap();
        wit_bindgen::spawn(async move {
            assert!(writer.write_all(data).await.is_empty());
            drop(writer);
            let trailers = Fields::new();
            trailers.append("x-guest-trailer", b"done").unwrap();
            trailer_writer.write(Ok(Some(trailers))).await.unwrap();
            assert!(transmitted.await.is_ok());
        });
        Ok(response)
    }
}
export!(Service);
