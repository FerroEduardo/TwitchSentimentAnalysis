import logging
import json
import threading
from typing import Generator, List
from confluent_kafka import Consumer, KafkaError
from model import Message

class KafkaConsumerService:
    def __init__(self, config: dict[str, any], shutdown_event: threading.Event, topic: str):
        self.logger = logging.getLogger(__name__)
        self.shutdown_event = shutdown_event
        self.topic = topic
        self.logger.info("Starting consumer")
        self.consumer = Consumer(config)

    def process_message(self, message: str) -> Message:
        try:
            return Message.from_dict(json.loads(message))
        except json.JSONDecodeError as e:
            raise Exception(f"Failed to decode JSON: {message}") from e

    def consume(self) -> Generator[List[Message], None, None]:
        try:
            self.logger.info(f"Subscribing to {self.topic}")
            self.consumer.subscribe([self.topic])

            while not self.shutdown_event.is_set():
                msgs = self.consumer.consume(num_messages=100, timeout=1.0)

                if msgs is None:
                    continue

                if not msgs:
                    continue

                messages = []
                for msg in msgs:
                    if msg is None:
                        continue

                    if msg.error():
                        if msg.error().code() == KafkaError._PARTITION_EOF:
                            self.logger.warning(f"{msg.topic()} [{msg.partition()}] reached end at offset {msg.offset()}")
                        else:
                            self.logger.error(msg.error())
                            continue
                    else:
                        try:
                            messages.append(self.process_message(msg.value().decode('utf-8')))
                        except Exception as e:
                            self.logger.info(e)
                            continue

                yield messages
                
        except Exception as e:
            self.logger.error(f"Error in consumer loop: {e}")
        finally:
            self.consumer.close()
            self.logger.info("Consumer closed")
