#!/bin/sh

CFG_MAIN=/etc/metcap/main.conf
CFG_MUT=/etc/metcap/graphite_mutator.conf

TMPL_MAIN=$( cat <<-EOF
###########################
# MetCap main config file #
###########################

syslog = false
debug = false
report_every = "${REPORT_EVERY}"

[transport]
type = "${TRANSPORT_TYPE}"
buffer_size = ${TRANSPORT_SIZE}
EOF
)

TMPL_REDIS=$( cat <<-EOF
redis_url = "${REDIS_URL}"
redis_timeout = ${REDIS_TIMEOUT}
redis_wait = ${REDIS_WAIT}
redis_retries = ${REDIS_RETRIES}
redis_connections = ${REDIS_CONNECTIONS}
redis_queue = "${REDIS_QUEUE}"
EOF
)

TMPL_AMQP=$( cat <<-EOF
amqp_url = "${AMQP_URL}"
amqp_timeout = ${AMQP_TIMEOUT}
amqp_tag = "${AMQP_TAG}"
amqp_workers = ${AMQP_WORKERS}
EOF
)

TMPL_LISTENER=$( cat <<-EOF

[listener]

  [listener.influx]
  port = 8001
  protocol = "tcp"
  codec = "influx"
  decoders = ${LISTENER_DECODERS}

  [listener.graphite]
  port = 8002
  protocol = "tcp"
  codec = "graphite"
  decoders = ${LISTENER_DECODERS}
  mutator_file = "${CFG_MUT}"
EOF
)

TMPL_WRITER=$( cat <<-EOF

[writer]
urls = [ "${WRITER_ES_URL}" ]
timeout = ${WRITER_TIMEOUT}
concurrency = ${WRITER_CONCURRENCY}
bulk_max = ${WRITER_BULK_MAX}
bulk_wait = "${WRITER_BULK_WAIT}"
index = "${WRITER_INDEX}"
doc_type = "${WRITER_DOC_TYPE}"
EOF
)

if [ ! -e ${CFG_MAIN} ]; then
  echo "${TMPL_MAIN}" >> ${CFG_MAIN}
  case "${TRANSPORT_TYPE}" in
    "redis")
      echo "${TMPL_REDIS}" >> ${CFG_MAIN}
      ;;
    "amqp")
      echo "${TMPL_AMQP}" >> ${CFG_MAIN}
      ;;
  esac

  case "${1}" in
    "listener")
      echo "${TMPL_LISTENER}" >> ${CFG_MAIN}
      shift
      ;;
    "writer")
      echo "${TMPL_WRITER}" >> ${CFG_MAIN}
      shift
      ;;
    *)
      echo "${TMPL_LISTENER}" >> ${CFG_MAIN}
      echo "${TMPL_WRITER}" >> ${CFG_MAIN}
      ;;
  esac

  echo "############### CONFIG END #################" >> ${CFG_MAIN}

  if [ ! -e ${CFG_MUT} ]; then
    echo ${LISTENER_GRAPHITE_RULES} > ${CFG_MUT}
  fi

fi

set -x
exec /bin/metcap $@
