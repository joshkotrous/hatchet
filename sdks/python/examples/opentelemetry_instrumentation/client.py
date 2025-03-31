from hatchet_sdk import Hatchet
import os

# Get debug setting from environment variables, default to False for security
debug_mode = os.environ.get('HATCHET_DEBUG', 'false').lower() == 'true'

hatchet = Hatchet(debug=debug_mode)