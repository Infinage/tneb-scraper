import os
import datetime as dt
from scraper import TNEBScraper
from utils import getLogger

if __name__ == "__main__":
    RETRY_ATTEMPTS = int(os.environ["RETRY_ATTEMPTS"])
    while RETRY_ATTEMPTS > 0:

        # Create a directory for debugging related info
        debug_path = f"./debug/{dt.datetime.today().isoformat().replace(':', '.')}"
        if not os.path.exists(debug_path):
          os.mkdir(debug_path)

        # Get the logger
        logger = getLogger(f"{debug_path}/tneb-scraper.log")

        try:
            script_runner = TNEBScraper(debug_path, logger)
            script_runner.execute()
        except Exception as e:
            logger.error(f"An error occurred: {e}")
            RETRY_ATTEMPTS -= 1
        else:
            logger.info("Job executed successfully.")
            break