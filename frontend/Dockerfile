# Build stage
FROM node:20-alpine AS builder

# Accept build argument for API target
ARG VITE_API_URL=http://localhost:18080
ENV VITE_API_URL=${VITE_API_URL}
ARG VITE_API_TARGET=http://localhost:18080
ENV VITE_API_TARGET=${VITE_API_TARGET}
ARG VITE_BACKEND_URL=http://localhost:18080
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
ARG VITE_GRAPH_SERVICE_URL=http://localhost:18080/graph-service
ENV VITE_GRAPH_SERVICE_URL=${VITE_GRAPH_SERVICE_URL}
ARG VITE_SKIP_AUTH=false
ENV VITE_SKIP_AUTH=${VITE_SKIP_AUTH}

WORKDIR /app

# Copy package files from frontend/
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci || npm install

# Copy frontend source code
COPY frontend/. .

# Copy config from build context for Vite alias $config
COPY knowledge-graph.config.json /knowledge-graph.config.json

# Build the application
RUN npm run build

# Remove sourcemap files to reduce size
RUN find build -name "*.map" -type f -delete && find . -type f -name '*.map' -delete

# Production stage
FROM node:20-alpine

WORKDIR /app

# Copy built application
COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json .

# Install curl for health check and production dependencies
RUN apk add --no-cache curl wget && \
    npm ci --omit=dev || npm install --omit=dev

EXPOSE 3000

ENV NODE_ENV=production
ENV PORT=3000

HEALTHCHECK --interval=30s --timeout=5s --retries=5 --start-period=10s \
  CMD wget -q -O- http://localhost:3000/ || exit 1

CMD ["node", "build"]
