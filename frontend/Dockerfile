# Build stage
FROM node:20-alpine AS builder

# Accept build argument for API target
ARG VITE_API_TARGET=http://localhost:8080
ENV VITE_API_TARGET=${VITE_API_TARGET}
ARG VITE_BACKEND_URL=http://localhost:8080
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
ARG VITE_GRAPH_SERVICE_URL=http://localhost:9091/graph-service
ENV VITE_GRAPH_SERVICE_URL=${VITE_GRAPH_SERVICE_URL}

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
RUN apk add --no-cache curl && \
    npm ci --omit=dev || npm install --omit=dev

EXPOSE 3000

ENV NODE_ENV=production
ENV PORT=3000

CMD ["node", "build"]
