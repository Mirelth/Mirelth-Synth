"""Synth Python SDK — v2.0"""

from .client import Synth
from .models import Agent, AgentConfig, Session, Tool, Message
from .exceptions import (
    SynthError, APIError, AuthenticationError,
    RateLimitError, TimeoutError, ValidationError
)

__version__ = "2.0.0"
__all__ = [
    "Synth", "Agent", "AgentConfig", "Session",
    "Tool", "Message",
    "SynthError", "APIError", "AuthenticationError",
    "RateLimitError", "TimeoutError", "ValidationError",
]

