import traceback


def errorWithTraceback(message: str, e: Exception, include_traceback=True):
    """
    Create an error message with an optional traceback.
    
    Args:
        message: The error message
        e: The exception that was caught
        include_traceback: Whether to include the full traceback in the message
    
    Returns:
        A formatted error message, with traceback if requested
    """
    if include_traceback:
        trace = "".join(traceback.format_exception(type(e), e, e.__traceback__))
        return f"{message}\n{trace}"
    else:
        # Return a sanitized message without sensitive traceback information
        return f"{message}\nError type: {type(e).__name__}\nError message: {str(e)}"