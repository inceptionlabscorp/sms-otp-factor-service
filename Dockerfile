FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY .build/server /server
USER nonroot:nonroot
ENTRYPOINT ["/server"]
