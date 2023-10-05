import crypto from "crypto";

export function sha256(input) {
  let hash = crypto.createHash("sha256");
  hash.update(input);
  return hash.digest();
}

export function sha256x2(input) {
  return sha256(sha256(input));
}
