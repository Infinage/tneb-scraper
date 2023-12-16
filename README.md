
## 🚀 Introduction

- Python selenium script developed to scrape TNEB website for electricity consumption charges.
- Scheduled to run every Sunday at 00:00 (refer to .github/workflows/schedule-scraper.yml).
- Utilizes `Tesseract` for bypassing the captcha validation during the scraping process.
- Implements automated retries in case of scraping failures due to captcha or other issues.
- Sends scraping results (success/failure) via email after each job run.

## 🛠️ Getting started

1. Clone the github repository
2. Ensure you have `docker` installed on your machine
3. Create a _.env_ file. Clone the _sample.env_ file and customize it according to your needs.
4. Create an _eb-mapping.json_ file containing mappings for the Consumer No. You can use the sample-mapping.json file as a reference.
5. Run command `docker compose up` (or `docker-compose up`)