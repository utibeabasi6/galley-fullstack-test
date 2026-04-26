# Static-only — no build step. Plain HTML + a single inline script.
# Kept deliberately tiny so this PR exercises Galley's "two web services
# in one preview" path without bloating build time.

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY index.html /usr/share/nginx/html/index.html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
