/**
 
3. In case of error, retry again after sometime. 
5. Provision to take screenshots of consumerNumber wise bills for reference & confirmation
6. Generate logs for job run / failure. The logs should be accessible by providing path along with credentials

 */

// Init all the confidential key-value pairs from .env
import { config } from "dotenv";
config();

// We use node-cron to schedule jobs
import * as cron from "node-cron";

// Used to prefix timestamps to console logs
import consoleStamp from "console-stamp";
consoleStamp(console);

// TNEB Scraping and Mailing logic begins here
import { getEbBills } from "./tneb-scraper"; 
import { sendEBBillAsMail } from "./tneb-mailer";

// Write to function to scrape and write email
const startJob = async (): Promise<void> => {
    const MAX_RETRY_ATTEMPTS = Number(process.env.RETRY_ATTEMPTS) || 3;
    let tryCount = 0;
    while (tryCount < MAX_RETRY_ATTEMPTS) {
        try {
            console.info(`Job starting @ ${new Date()}`);
            const ebBills = await getEbBills();
            await sendEBBillAsMail(ebBills);
            tryCount = MAX_RETRY_ATTEMPTS;
            console.info(`Job complete @ ${new Date()}`);
        } catch (err) {
            console.error(`An exception has occured: ${err} Retrying again - #${++tryCount}.`);
        }
    }
}

// Schedule cron job to run at 10th of each month by 8am
console.info(`The Application is up. And is scheduled to run ${process.env.CRON_EXP_DESC}`);
cron.schedule(`${process.env.CRON_EXP}`, () => startJob());