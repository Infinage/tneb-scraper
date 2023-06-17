import * as nodemailer from "nodemailer";
import { ebBillType } from "./tneb-scraper";

export const sendEBBillAsMail = async (bills: ebBillType[]): Promise<void> => {

    try {

        if (bills.length > 0){
            const transporter = nodemailer.createTransport({
                service: "gmail",
                auth: {
                    user: process.env.GMAIL_FROM_ADDRESS,
                    pass: process.env.GMAIL_APP_PWD
                }
            });
            
            const mailOptions = {
                from: process.env.GMAIL_FROM_ADDRESS,
                to: process.env.GMAIL_TO_ADDRESS,
                subject: `${new Date().toISOString().slice(0, 10)} | EB Bills Scraped Results`,
                html: ebBillsToHtml(bills)
            };
        
            const mailSentInfo = await transporter.sendMail(mailOptions);
            console.info('Mail sent successfully: %s', mailSentInfo.messageId);
        } else {
            throw new Error("No scraped results were found.");
        }

    } catch (err) {
        console.error(`An error had occured when sending mail: ${err}`);
    }
}

// Helper function that iterates through bill array and parses it to a HTML Table
const ebBillsToHtml = (bills: ebBillType[]) => {

    console.debug(`Writing scraped results to HTML. Total Bill Count: ${bills.length}`)
    
    let message = ('<table><thead><tr>' + 
    '<th>S No.</th>' + 
    '<th>Consumer Number</th>' + 
    '<th>Portion Code</th>' + 
    '<th>Bill Amount</th>' + 
    '<th>Due Date</th>' + 
    '</tr></thead><tbody>');

    for (let [index, bill] of bills.entries()){
        message = message + (`<tr>` + 
        `<td>${index}</td>` + 
        `<td>${bill.consumerNumber}</td>` + 
        `<td>${bill.portionCode}</td>` + 
        `<td>${bill.billAmount}</td>` + 
        `<td>${bill.dueDate}</td></tr>`);

        console.debug(`Successfuly parsed #${index + 1} to HTML.`);
    }
    
    message = message + '</tbody></table>';
    return message;

}