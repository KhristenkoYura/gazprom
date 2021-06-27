FROM frapsoft/openssl

ADD main /app/

WORKDIR /
EXPOSE 8080
EXPOSE 6060

ENTRYPOINT [ "/app/main" ]