import * as puppeteer from "puppeteer";
import ebMapping from "./eb-mapping";
import { extractText } from "./extract-capcha";

export type ebBillType = {
    consumerNumber: String,
    portionCode: String,
    billAmount: String,
    dueDate: String
}

export const getEbBills = async () => {

    let ebBills: ebBillType[] = [];

    const browser = await puppeteer.launch({ 
        headless: "new", ignoreHTTPSErrors: true,
        executablePath: puppeteer.executablePath(),
        args: ['--no-sandbox','--disable-setuid-sandbox']
    });

    console.info("Opened puppeteer browser.");
    
    const page = await browser.newPage();

    console.info("Opened a new blank page.");

    try {

        if (!process.env.TNEB_LOGIN_URL || !process.env.TNEB_USERNAME || !process.env.TNEB_PASSWORD) {
            throw new Error("Environment variables not found.");
        }

        await page.goto(process.env.TNEB_LOGIN_URL);

        console.debug("Finished opening TANGEDCO Login Page.");

        // Save the captcha image to get the text
        const captchaImageElement: puppeteer.ElementHandle | null = await page.$("img#CaptchaImgID");
        const captchaScreenshotPath: string = `./debug/captcha-${new Date().toISOString()}.png`
        if (captchaImageElement) {
            await captchaImageElement.screenshot({path: captchaScreenshotPath});
            console.debug(`Screenshot of Captcha was taken. Processing with Tesseract.`);
        } else throw new Error(`Captcha screenshot couldn't be taken`);

        const captcha: string = await extractText(captchaScreenshotPath);
        console.debug(`Captcha is: ${captcha.trim()}`);

        // Type in the password and captcha
        await page.$eval('input#userName', (el, value) => el.value = value, process.env.TNEB_USERNAME || "");
        await page.$eval('input#password', (el, value) => el.value = value, process.env.TNEB_PASSWORD || "");
        await page.$eval('input#CaptchaID', (el, value) => el.value = value, captcha);

        await Promise.all([
            page.waitForNavigation(),
            page.click("input[name='submit'][value='Login']"),
        ]);

        console.info("Login to TANGEDCO successful.");

        let scrapedElements: puppeteer.ElementHandle<HTMLDivElement>[] = [];
        let scrapedData: (String | null)[] = [], pageCount = 1;

        scrapedElements.push(...await page.$$("tr :not(:first-child) td div.ui-dt-c"));
        scrapedData = await Promise.all(scrapedElements.map(async ele => await (await ele.getProperty("textContent")).jsonValue()));
        console.debug(`Scraped from Page# ${pageCount++}: ${scrapedData.length / 6} consumer numbers so far.`);
        console.debug("Finished Scraping from the Web Page.");

        for (let i = 0; i < scrapedData.length / 6; i++){
            
            let consumerNumber = (scrapedData[i * 6]) as string;
            let billAmount = scrapedData[i * 6 + 3];
            let dueDate = scrapedData[i * 6 + 4];

            if (consumerNumber && billAmount && Number.isInteger(Number(billAmount.slice(3)))) {
                ebBills.push({
                    portionCode: ebMapping.get(consumerNumber) || "Unknown",
                    consumerNumber: consumerNumber,
                    billAmount: billAmount,
                    dueDate: dueDate || "-"
                })
            }
        }

        console.info("Scraped results have been successfully parsed");

    } catch(err) {
        const screenShotPath = `./debug/screenshot-${new Date().toISOString()}.png`;
        console.error(`An error had occured when scraping: ${err}. Saving screenshot.`);
        try {
            await page.screenshot({ path: screenShotPath });
            console.debug(`A screenshot has been saved to path: ${screenShotPath}.`);
        } catch (screenshotError) {
            console.error(`An error has occured in saving screenshot: ${screenshotError}`)
        }
    } finally {
        browser.close();
        console.debug(`Browser was closed. ${ebBills.length} rows have been scraped in total.`);
        return ebBills;
    }
}