import os
from scraper import TNEBScraper

if __name__ == "__main__":
    RETRY_ATTEMPTS = int(os.environ["RETRY_ATTEMPTS"])
    while RETRY_ATTEMPTS > 0:
        try:
            script_runner = TNEBScraper()
            script_runner.execute()
        except Exception as e:
            print(f"An error occurred: {e}")
            RETRY_ATTEMPTS -= 1
        else:
            print("Job executed successfully.")
            break