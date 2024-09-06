#syntax=docker/dockerfile:1.9.0

FROM alpine AS prepare
RUN <<EOF
echo "bar" >/foo
EOF

FROM scratch
COPY --from=prepare /foo /foo