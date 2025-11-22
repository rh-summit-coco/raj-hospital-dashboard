FROM registry.redhat.io/ubi9/nginx-120:latest

# Copy the dashboard HTML file
COPY index.html /opt/app-root/src/index.html

# Create a simple nginx config
USER root
RUN echo 'server {' > /etc/nginx/nginx.conf && \
    echo '    listen 8080;' >> /etc/nginx/nginx.conf && \
    echo '    location / {' >> /etc/nginx/nginx.conf && \
    echo '        root /opt/app-root/src;' >> /etc/nginx/nginx.conf && \
    echo '        index index.html;' >> /etc/nginx/nginx.conf && \
    echo '        try_files $uri $uri/ /index.html;' >> /etc/nginx/nginx.conf && \
    echo '    }' >> /etc/nginx/nginx.conf && \
    echo '}' >> /etc/nginx/nginx.conf && \
    echo 'events { worker_connections 1024; }' >> /etc/nginx/nginx.conf

USER 1001

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]