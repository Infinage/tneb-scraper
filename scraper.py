from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
import pytesseract, json, os
import datetime as dt
import pandas as pd
from io import StringIO
from mailer import TNEBMailer
import logging

class TNEBScraper:

    def __init__(self, debug_path: str, logger: logging.Logger):

      # Load from .env to our custom class
      self.FROM_EMAIL = os.environ["GMAIL_FROM_ADDRESS"]
      self.TO_EMAIL = os.environ["GMAIL_TO_ADDRESS"]
      self.FROM_EMAIL_PWD = os.environ["GMAIL_APP_PWD"]
      self.TNEB_LOGIN_URL = os.environ["TNEB_LOGIN_URL"]
      self.TNEB_PASSWORD = os.environ["TNEB_PASSWORD"]
      self.TNEB_USERNAME = os.environ["TNEB_USERNAME"]
      self.RETRY_ATTEMPTS = os.environ["RETRY_ATTEMPTS"]

      # Initialize the Mailer utility
      self.mailer = TNEBMailer(
          logger=logger, debug_path=debug_path, 
          FROM_EMAIL=self.FROM_EMAIL, FROM_EMAIL_PWD=self.FROM_EMAIL_PWD, 
          TO_EMAIL=self.TO_EMAIL
      )

      # Path to log files and save screenshots
      self.debug_path = debug_path

      # Logger object to log to console and to a file
      self.logger = logger

      # Begin using our logger
      self.logger.info("Environment variables loaded successfully.")

    def execute(self):
        
        # Scrape the bills
        bill_details = self.scrape_bills()

        # Send the mail with processed details
        self.mailer.send_mail(bill_details)

        # If bill_details is none, there had been some error - raise exeception to trigger a retry
        if bill_details is None:
            raise Exception("Debug mail sent but there had been some error.")

    def scrape_bills(self) -> pd.DataFrame:

        # Open Chrome 
        chrome_options = Options()
        chrome_options.add_argument("--headless")
        chrome_options.add_argument("--no-sandbox")
        chrome_options.add_argument("--disable-gpu")
        chrome_options.add_argument('--start-maximized')
        driver = webdriver.Chrome(options=chrome_options)
        self.logger.info("Chrome opened successfully.")

        try:

          # open TNEB Login page
          driver.get(self.TNEB_LOGIN_URL)
          driver.save_screenshot(f"{self.debug_path}/login-page.png")
          self.logger.info("TNEB Page opened successfully.")

          # Input username
          input_elem = driver.find_element(By.CSS_SELECTOR, "input#userName")
          input_elem.clear()
          input_elem.send_keys(self.TNEB_USERNAME)
          self.logger.info("Username input successfully.")

          # Input password
          pwd_elem = driver.find_element(By.CSS_SELECTOR, "input#password")
          pwd_elem.clear()
          pwd_elem.send_keys(self.TNEB_PASSWORD)
          self.logger.info("Password input successfully.")

          # Scrape the captcha text
          captcha_img_path = f"{self.debug_path}/captcha.png"
          driver.find_element(By.CSS_SELECTOR, "img#CaptchaImgID").screenshot(captcha_img_path)
          captcha_text = pytesseract.image_to_string(captcha_img_path, lang="eng", config='--psm 8 -c tessedit_char_whitelist=0123456789').strip()
          self.logger.info(f"Scraped image and processed it successfully: {captcha_text}")

          # Input scraped captcha
          captcha_inp_elem = driver.find_element(By.CSS_SELECTOR, "input#CaptchaID")
          captcha_inp_elem.clear()
          captcha_inp_elem.send_keys(captcha_text)
          driver.save_screenshot(f"{self.debug_path}/credentials-entered.png")
          self.logger.info("Scraped Captcha text input successfully.")

          # Enter the login button
          driver.find_element(By.CSS_SELECTOR, "input[name='submit'][value='Login']").click()
          self.logger.info("Submit button clicked.")

          # Search for `Bill Payments` legend and its sibling table
          bill_payments_legend_elem = driver.find_element(By.XPATH, "//legend[contains(text(), 'Bill Payments')]")
          bill_details_table_elem = bill_payments_legend_elem.find_element(By.XPATH, "//preceding::fieldset/descendant::table")
          bill_details_tbl_html = StringIO(bill_details_table_elem.get_attribute("outerHTML"))
          driver.save_screenshot(f"{self.debug_path}/bill-details.png")
          self.logger.info("Bill details HTML scraped successfully.")

        except Exception as e:
          # Log an error message, save screenshot
          self.logger.error(f"An error occurred while scraping TNEB: {e}")
          driver.save_screenshot(f"{self.debug_path}/error.png")
          return None
        
        else:
          # Parse the html with pandas and do some cleaning 
          bill_details = pd.read_html(bill_details_tbl_html)[0]
          self.logger.info(f"Bill details HTML parsed to DataFrame. Shape: {bill_details.shape}")

          bill_details.columns = bill_details.columns.map(lambda x: x[1])
          bill_details = bill_details.drop(["Select All", "Consumer Name"], axis=1)
          bill_details = bill_details.dropna(how='all')
          self.logger.info("Columns renamed & NA values were dropped.")

          # Load eb-mapping.json file
          eb_mapping = json.load(open("./eb-mapping.json"))
          self.logger.info("Loaded `eb-mapping.json` file as dictionary.")
          
          # Merge it over eb-mapping to get the associated Portion
          eb_mapping = pd.Series(eb_mapping).reset_index().rename({"index": "Consumer No", 0: "Portion"}, axis=1)
          eb_mapping['Consumer No'] = eb_mapping['Consumer No'].astype('int64')
          bill_details = bill_details.merge(eb_mapping, on='Consumer No').sort_values("Portion")
          bill_details = bill_details.reset_index(drop=True)
          self.logger.info(f"Concatenated with `eb-mapping.json` file. Shape: {bill_details.shape}")

          # Return the dataframe for downstream ops
          return bill_details