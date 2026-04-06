FROM ghcr.io/cirruslabs/flutter:3.22.0 AS builder
WORKDIR /app
COPY . .
RUN flutter pub get
RUN flutter build web

FROM nginx:alpine
# Copy compiled flutter web app
COPY --from=builder /app/build/web /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
