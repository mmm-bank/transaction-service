FROM postgres:15

RUN apt-get update && apt-get install -y \
    build-essential \
    postgresql-server-dev-15 \
    git \
    libxml2-dev \
    libxslt-dev \
    && rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/omniti-labs/pg_amqp.git /usr/src/pg_amqp

WORKDIR /usr/src/pg_amqp
RUN make && make install