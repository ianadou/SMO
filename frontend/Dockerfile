# SMO frontend — Nuxt 3 SPA
#
# Stage 1 (builder): full Node + pnpm to install deps and run the
# Nuxt production build. The output goes to /app/.output, which is a
# self-contained Nitro server bundling all static assets.
#
# Stage 2 (runtime): minimal Node runtime that serves the bundle.
# Distroless is NOT used here: unlike the Go binary, Nuxt needs the
# Node runtime to execute the Nitro server, and node:24-alpine is
# already small (~50 MB) while still letting wget answer the health
# probe.

# ----- Stage 1: builder ------------------------------------------------------
FROM node:24.10.0-alpine AS builder

# Pnpm via corepack — the version is pinned by the project's
# packageManager field in package.json once set, so corepack picks
# the right one automatically.
RUN corepack enable

WORKDIR /app

# Install dependencies first to maximise cache reuse: re-running
# this layer only happens when the lockfile changes, not on every
# source edit.
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Now bring the rest of the source and build.
COPY . .
RUN pnpm build

# ----- Stage 2: runtime ------------------------------------------------------
FROM node:24.10.0-alpine AS runtime

WORKDIR /app

# Copy only the build output. node_modules is bundled inside .output
# by Nitro, so no separate install step is needed at runtime.
COPY --from=builder /app/.output ./.output

# Run as the non-root user that the official node image ships with
# (uid 1000). The Nitro server writes nothing to the working dir at
# runtime, so /app stays read-only-friendly for free.
USER node

ENV HOST=0.0.0.0
ENV PORT=3000
EXPOSE 3000

# Healthcheck: hit the root, the SPA index always returns 200 once
# Nitro is ready. wget is included in node:alpine.
HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=3 \
  CMD wget -qO- "http://127.0.0.1:${PORT}/" >/dev/null 2>&1 || exit 1

CMD ["node", ".output/server/index.mjs"]
