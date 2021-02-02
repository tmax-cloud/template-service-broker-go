FROM python:alpine

RUN mkdir -p /devpi-home

WORKDIR /devpi-home

RUN apk add --no-cache --virtual .build-deps gcc libffi-dev musl-dev supervisor && \
    pip install -q -U devpi-server && \
	rm -r /root/.cache

COPY serve.sh /serve.sh

ENTRYPOINT ["/serve.sh"]

