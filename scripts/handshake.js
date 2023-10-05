import net from "net";
import crypto from "crypto";
import { sha256x2 } from "./utils.js";

const PEER_HOST = "35.182.93.221";
const PEER_PORT = "8333";

const MAX_HEADER_BYTES = 24;

const payload = Buffer.alloc(1024);
payload.writeUInt32LE(70001, 0, 16); // Version
payload.writeDoubleLE(0, 4, 16); // Services
payload.writeDoubleLE(Math.floor(Date.now() / 1000), 8); // Timestamp
payload.write("", 12, 16); // Addr_recv
payload.write("", 38, 16); // Addr_from
payload.writeUInt8(crypto.randomBytes(8), 52); // Nonce

const header = Buffer.alloc(MAX_HEADER_BYTES);
header.writeUInt32LE(0xd9b4bef9, 0); // Magic
header.write("version", 4, 16); // Command
header.writeUInt32LE(payload.length, 16); // Payload length
header.writeUInt32LE(sha256x2(payload).readUInt32LE(0), 20); // Checksum

const message = Buffer.concat(
  [header, payload],
  MAX_HEADER_BYTES + payload.length
);

console.log("Connecting...");

const client = net.createConnection(
  {
    host: PEER_HOST,
    port: PEER_PORT,
  },
  () => {
    console.log("Connected");
    client.write(message);
  }
);

client.on("data", (data) => {
  console.log("Data:", data, data.length);
});

client.on("error", (err) => {
  console.log(`Error: ${err.message}`);
});

client.on("end", () => {
  console.log("Disconnected");
});
