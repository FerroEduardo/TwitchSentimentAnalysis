import os
import signal
import sys
import logging
import threading
from kafka import KafkaConsumerService
from analyzer import SentimentAnalyzer
from database import DatabaseWriter

logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[logging.StreamHandler(sys.stdout)]
)

logger = logging.getLogger(__name__)
shutdown_event = threading.Event()

def shutdown(signum, frame):
    global shutdown_event
    logger.info("Shutting down...")
    shutdown_event.set()

signal.signal(signal.SIGINT, shutdown)
signal.signal(signal.SIGTERM, shutdown)

if __name__ == "__main__":
    try:
        service = KafkaConsumerService(config={
            'bootstrap.servers': os.getenv("KAFKA_BROKER_HOST"),
            'group.id': 'analyzer',
            'auto.offset.reset': 'smallest'
        }, topic="messages", shutdown_event=shutdown_event)
        analyzer = SentimentAnalyzer()
        writer = DatabaseWriter({
            'dbname': os.getenv("DATABASE_DATABASE"),
            'user': os.getenv("DATABASE_USER"),
            'password': os.getenv("DATABASE_PASSWORD"),
            'host': os.getenv("DATABASE_HOST"),
            'port': os.getenv("DATABASE_PORT")
        })
        for messages in service.consume():
            logger.info(f"Received {len(messages)} messages")
            result = analyzer.analyze([msg.message for msg in messages])
            writer.insert_results(
                [
                {
                    'channel': messages[i].channel,
                    'user': messages[i].user,
                    'id': messages[i].id,
                    'timestamp': messages[i].timestamp,
                    'message': messages[i].message,
                    'sentiment_positive': result[i].positive,
                    'sentiment_neutral': result[i].neutral,
                    'sentiment_negative': result[i].negative,
                }
                for i in range(len(messages))
            ]
            )
    except Exception as e:
        logger.error(f"Fatal error: {e}")
    finally:
        writer.close()
        logger.info("Exiting")
