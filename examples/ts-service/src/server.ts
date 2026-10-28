import { createServer, type IncomingMessage, type Server, type ServerResponse } from "node:http";
import { quote, type LineItem } from "./pricing.ts";

const MAX_BODY_BYTES = 64 * 1024;

function send(res: ServerResponse, status: number, body: unknown): void {
  res.writeHead(status, { "content-type": "application/json", "x-content-type-options": "nosniff" });
  res.end(JSON.stringify(body));
}

async function readJSON(req: IncomingMessage): Promise<unknown> {
  const chunks: Buffer[] = [];
  let size = 0;
  for await (const chunk of req) {
    const buf = chunk as Buffer;
    size += buf.length;
    if (size > MAX_BODY_BYTES) {
      throw new RangeError("request body too large");
    }
    chunks.push(buf);
  }
  return JSON.parse(Buffer.concat(chunks).toString("utf8"));
}

export function buildServer(): Server {
  return createServer((req, res) => {
    if (req.method === "GET" && req.url === "/healthz") {
      send(res, 200, { status: "ok" });
      return;
    }
    if (req.method === "POST" && req.url === "/quote") {
      readJSON(req)
        .then((body) => {
          const items = (body as { items?: LineItem[] }).items;
          if (!Array.isArray(items)) {
            send(res, 400, { error: "items must be an array" });
            return;
          }
          send(res, 200, quote(items));
        })
        .catch((err: unknown) => {
          send(res, 400, { error: err instanceof Error ? err.message : "bad request" });
        });
      return;
    }
    send(res, 404, { error: "not found" });
  });
}
