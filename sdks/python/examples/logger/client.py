# ❓ RootLogger


import logging
import os

from hatchet_sdk import ClientConfig, Hatchet

logging.basicConfig(level=logging.INFO)

root_logger = logging.getLogger()

# Use environment variable to control debug mode, defaults to False for security
debug_mode = os.environ.get('HATCHET_DEBUG', '').lower() in ('true', 'yes', '1')

hatchet = Hatchet(
    debug=debug_mode,
    config=ClientConfig(
        logger=root_logger,
    ),
)

# ‼️