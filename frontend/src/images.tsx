import { useEffect, useState } from "react";
import { request } from "./api";
// Referenced URLs cannot be evicted. Unused entries are kept under a byte budget.
interface Entry {
  url?: string;
  bytes: number;
  refs: number;
  used: number;
  promise: Promise<string>;
  controller: AbortController;
}
const entries = new Map<string, Entry>();
const MAX_BYTES = 48 * 1024 * 1024;
function trim() {
  let bytes = [...entries.values()].reduce((n, e) => n + e.bytes, 0);
  for (const [key, e] of [...entries].sort((a, b) => a[1].used - b[1].used)) {
    if (bytes <= MAX_BYTES && entries.size <= 160) break;
    if (e.refs || !e.url) continue;
    URL.revokeObjectURL(e.url);
    entries.delete(key);
    bytes -= e.bytes;
  }
}
export function clearImages() {
  for (const e of entries.values()) {
    e.controller.abort();
    if (e.url) URL.revokeObjectURL(e.url);
  }
  entries.clear();
}
function acquire(path: string) {
  let entry = entries.get(path);
  if (!entry) {
    const created: Entry = {
      controller: new AbortController(),
      bytes: 0,
      refs: 0,
      used: Date.now(),
      promise: Promise.resolve(""),
    };
    created.promise = request(path, created.controller.signal)
      .then((r) => r.blob())
      .then((blob) => {
        if (!blob.type.startsWith("image/"))
          throw new Error("Image indisponible.");
        const url = URL.createObjectURL(blob);
        if (entries.get(path) !== created) {
          URL.revokeObjectURL(url);
          throw new Error("Session terminée.");
        }
        created.url = url;
        created.bytes = blob.size;
        trim();
        return url;
      })
      .catch((error) => {
        if (entries.get(path) === created) entries.delete(path);
        throw error;
      });
    entries.set(path, created);
    entry = created;
  }
  entry.refs++;
  entry.used = Date.now();
  const held = entry;
  return {
    promise: entry.promise,
    release: () => {
      held.refs--;
      held.used = Date.now();
      if (!held.refs && !held.url) {
        held.controller.abort();
        if (entries.get(path) === held) entries.delete(path);
      }
      trim();
    },
  };
}
export function useImage(path?: string) {
  const [state, setState] = useState<{
    path?: string;
    url?: string;
    error?: string;
  }>({});
  const [retry, setRetry] = useState(0);
  useEffect(() => {
    if (!path) return;
    let active = true;
    const held = acquire(path);
    held.promise.then(
      (url) => {
        if (active) setState({ path, url });
      },
      () => {
        if (active) setState({ path, error: "Image indisponible" });
      },
    );
    return () => {
      active = false;
      held.release();
    };
  }, [path, retry]);
  return {
    ...(state.path === path ? state : {}),
    retry: () => {
      setState({});
      setRetry((n) => n + 1);
    },
  };
}
export function PhotoImage({
  path,
  alt,
  className = "",
}: {
  path: string;
  alt: string;
  className?: string;
}) {
  const { url, error, retry } = useImage(path);
  if (error)
    return (
      <span
        className="image-error"
        role="img"
        aria-label={error}
        onClick={(e) => {
          e.stopPropagation();
          retry();
        }}
      >
        Image indisponible · Réessayer
      </span>
    );
  return url ? (
    <img className={className} src={url} alt={alt} draggable={false} />
  ) : (
    <span
      className={`image-placeholder ${className}`}
      aria-label="Chargement de la photo"
    />
  );
}
