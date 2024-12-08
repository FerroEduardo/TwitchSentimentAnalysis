from dataclasses import dataclass
from typing import Any, Dict

@dataclass
class Message:
    id: str
    message: str
    channel: str
    user: str
    timestamp: int

    @staticmethod
    def from_dict(data: dict[str, Any]) -> "Message":
        return Message(
            id=data["id"],
            message=data["message"],
            channel=data["channel"],
            user=data["user"],
            timestamp=data["timestamp"]
        )

    def to_dict(self) -> dict[str, Any]:
        return {
            "id": self.id,
            "message": self.message,
            "channel": self.channel,
            "user": self.user,
            "timestamp": self.timestamp
        }
    
    def __str__(self) -> str:
        return (
            f"Message(id='{self.id}', "
            f"message='{self.message}', "
            f"channel='{self.channel}', "
            f"user='{self.user}', "
            f"timestamp={self.timestamp})"
        )

@dataclass
class SentimentResult:
    negative: float
    positive: float
    neutral: float

    @staticmethod
    def from_dict(data: Dict[str, Any]) -> "SentimentResult":
        fields = {item['label']: item['score'] for item in data}
        return SentimentResult(
            negative=fields.get('negative', 0.0),
            positive=fields.get('positive', 0.0),
            neutral=fields.get('neutral', 0.0)
        )

    def to_dict(self) -> Dict[str, float]:
        return {
            'negative': self.negative,
            'positive': self.positive,
            'neutral': self.neutral
        }

    def __str__(self) -> str:
        return (
            f"SentimentResult(negative={self.negative:.4f}, "
            f"positive={self.positive:.4f}, neutral={self.neutral:.4f})"
        )