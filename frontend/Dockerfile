# Build stage
FROM node:20-alpine AS builder

# Accept build argument for API target
ARG VITE_API_TARGET=http://localhost:8080
ENV VITE_API_TARGET=${VITE_API_TARGET}

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

# Production stage
FROM node:20-alpine

WORKDIR /app

# Copy built application
COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json .

# Install only production dependencies
RUN npm ci --omit=dev || npm install --omit=dev

EXPOSE 3000

ENV NODE_ENV=production
ENV PORT=3000

CMD ["node", "build"]
