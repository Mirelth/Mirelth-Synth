class SynthError(Exception):
    """Base exception for all SDK errors."""
    def __init__(self, message: str, status_code: int | None = None, response=None):
        super().__init__(message)
        self.status_code = status_code
        self.response = response

class APIError(SynthError):
    """HTTP API returned non-2xx status."""
    pass

class AuthenticationError(SynthError):
    """Invalid or missing API key."""
    def __init__(self, message="Authentication failed. Check your API key.", **kw):
        super().__init__(message, status_code=401, **kw)

class RateLimitError(SynthError):
    """Too many requests."""
    def __init__(self, message="Rate limit exceeded. Try again later.", **kw):
        super().__init__(message, status_code=429, **kw)

class TimeoutError(SynthError):
    """Request timed out."""
    def __init__(self, message="Request timed out.", **kw):
        super().__init__(message, **kw)

class ValidationError(SynthError):
    """Invalid request parameters."""
    def __init__(self, message="Validation failed.", **kw):
        super().__init__(message, status_code=422, **kw)
