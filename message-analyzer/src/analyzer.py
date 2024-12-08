from transformers import pipeline
from model import SentimentResult
from typing import List
import logging
import time

class SentimentAnalyzer:
    def __init__(self, model_name: str = "lxyuan/distilbert-base-multilingual-cased-sentiments-student", device: str = "cpu"):
        self.logger = logging.getLogger(__name__)
        self.model_name = model_name
        self.device = device
        self.logger.info(f"Starting model: {model_name}")
        self.sentiment_pipeline = pipeline(
            model=self.model_name,
            top_k=None,
            device=self.device
        )
        self.logger.info(f"Model started")

    def analyze(self, texts: List[str]) -> List[SentimentResult]:
        self.logger.debug(f"Analysing {len(texts)} texts")
        start = time.time()
        results = self.sentiment_pipeline(texts)
        end = time.time() - start
        self.logger.debug(f"Elapsed {end}s to process {len(texts)} items")

        return [
            SentimentResult.from_dict(result)
            for result in results
        ]