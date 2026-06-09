FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Все URL-ы указывают через nginx (:3000), а не напрямую на сервисы
ARG VITE_AUTH_ENABLED=true
ARG VITE_AUTH_SERVICE_URL=http://localhost:3000
ARG VITE_FLOWCHART_SERVICE_URL=http://localhost:3000
ARG VITE_METRICS_SERVICE_URL=http://localhost:3000
ARG VITE_INTERPRETER_SERVICE_URL=http://localhost:3000

ENV VITE_AUTH_ENABLED=$VITE_AUTH_ENABLED
ENV VITE_AUTH_SERVICE_URL=$VITE_AUTH_SERVICE_URL
ENV VITE_FLOWCHART_SERVICE_URL=$VITE_FLOWCHART_SERVICE_URL
ENV VITE_METRICS_SERVICE_URL=$VITE_METRICS_SERVICE_URL
ENV VITE_INTERPRETER_SERVICE_URL=$VITE_INTERPRETER_SERVICE_URL

RUN npm run build

FROM nginx:1.27-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]