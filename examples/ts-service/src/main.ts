import { buildServer } from "./server.ts";

const port = Number(process.env.PORT ?? 8080);
const server = buildServer();
server.listen(port, () => {
  console.log(JSON.stringify({ msg: "listening", port }));
});

for (const signal of ["SIGINT", "SIGTERM"] as const) {
  process.on(signal, () => server.close(() => process.exit(0)));
}
