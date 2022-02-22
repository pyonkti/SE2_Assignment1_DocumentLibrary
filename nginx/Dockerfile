FROM alpine:latest
COPY nginx.conf /etc/nginx/
RUN apk update && \
        apk --no-cache add nginx
EXPOSE 80
CMD ["nginx"]
