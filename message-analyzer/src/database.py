import psycopg2
from psycopg2 import sql
from psycopg2.extras import execute_values
from typing import List, Dict
import logging

class DatabaseWriter:
    def __init__(self, db_config: Dict[str, str]):
        self.db_config = db_config
        self.connection = None
        self.cursor = None
        self.logger = logging.getLogger(__name__)
        self.connect()

    def connect(self):
        try:
            self.connection = psycopg2.connect(**self.db_config)
            self.cursor = self.connection.cursor()
            self.logger.info("Connected to the database.")
        except Exception as e:
            self.logger.error(f"Error connecting to the database: {e}")
            raise

    def close(self):
        if self.cursor:
            self.cursor.close()
        if self.connection:
            self.connection.close()
        self.logger.info("Connection closed.")

    def insert_results(self, results: List[Dict]):
        self.logger.info(f"Inserting {len(results)} items to database.")
        insert_query = sql.SQL("""
            INSERT INTO results (
                channel, "user", message_id, "timestamp", message, 
                sentiment_positive, sentiment_neutral, sentiment_negative
            ) VALUES %s
        """)

        values = [
            (
                result['channel'],
                result['user'],
                result['id'],
                result['timestamp'],
                result['message'],
                result['sentiment_positive'],
                result['sentiment_neutral'],
                result['sentiment_negative']
            )
            for result in results
        ]

        try:
            execute_values(self.cursor, insert_query, values, template='(%s, %s, %s, to_timestamp(%s), %s, %s, %s, %s)')
            self.connection.commit()
            self.logger.info(f"Inserted {len(results)} records.")
        except Exception as e:
            self.logger.error(f"Error inserting multiple records: {e}")
            self.connection.rollback()
