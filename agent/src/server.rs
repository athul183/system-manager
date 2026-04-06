use axum::{
    extract::{ws::{Message, WebSocket, WebSocketUpgrade}, State},
    response::IntoResponse,
    routing::get,
    Router,
};
use tokio::sync::broadcast;
use crate::models::{AppMetrics, ServerPayload};

#[derive(Clone)]
struct AppState {
    tx: broadcast::Sender<Vec<AppMetrics>>,
}

pub async fn start_server(tx: broadcast::Sender<Vec<AppMetrics>>) {
    let state = AppState { tx };

    let app = Router::new()
        .route("/ws", get(ws_handler))
        .with_state(state);

    let listener = tokio::net::TcpListener::bind("0.0.0.0:8080").await.unwrap();
    println!("WebSocket server listening on 0.0.0.0:8080");
    axum::serve(listener, app).await.unwrap();
}

async fn ws_handler(
    ws: WebSocketUpgrade,
    State(state): State<AppState>,
) -> impl IntoResponse {
    ws.on_upgrade(|socket| handle_socket(socket, state))
}

async fn handle_socket(mut socket: WebSocket, state: AppState) {
    let mut rx = state.tx.subscribe();
    
    // We'll just stream data continuously
    while let Ok(metrics) = rx.recv().await {
        let payload = ServerPayload {
            msg_type: "metrics".to_string(),
            payload: metrics,
        };
        
        if let Ok(json) = serde_json::to_string(&payload) {
            if socket.send(Message::Text(json)).await.is_err() {
                break; // Client disconnected
            }
        }
    }
}
