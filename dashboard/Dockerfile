# Build Stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy dependency files first for caching
COPY package.json package-lock.json ./
RUN npm install

# Copy the source code
COPY . .

# Build the Next.js app
# Ensure we disable lint/typecheck during build for speed/resilience if needed
RUN npm run build

# Production Stage
FROM node:20-alpine AS runner

WORKDIR /app

# Copy the build artifacts and node_modules
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json

# Set environment defaults
ENV NODE_ENV=production
ENV PORT=3000

EXPOSE 3000

CMD ["npm", "start"]
