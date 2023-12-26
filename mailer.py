import pandas as pd
import datetime as dt
from email import encoders
from email.mime.multipart import MIMEMultipart
from email.mime.base import MIMEBase
from email.mime.text import MIMEText
from email.mime.image import MIMEImage
import logging, ssl, smtplib, os, zipfile, io, glob

class TNEBMailer:

    def __init__(self, logger: logging.Logger, debug_path: str, FROM_EMAIL: str, FROM_EMAIL_PWD: str, TO_EMAIL: str):
        self.logger = logger
        self.debug_path = debug_path
        self.FROM_EMAIL = FROM_EMAIL
        self.TO_EMAIL = TO_EMAIL
        self.FROM_EMAIL_PWD = FROM_EMAIL_PWD

    def send_mail(self, bill_details: pd.DataFrame):

        server = smtplib.SMTP("smtp.gmail.com", 587)
        context = ssl.create_default_context()

        try:
            if bill_details is None:

                self.logger.warn("TNEB could not be scraped.")

                # Create zipfile for attaching the logs
                zip_buffer = io.BytesIO()
                with zipfile.ZipFile(zip_buffer, 'a', zipfile.ZIP_DEFLATED, False) as zf:
                    for file in glob.glob(self.debug_path + "/*"):
                        with open(file, 'rb') as f:
                            zf.writestr(file.split(os.path.sep)[-1], f.read())

                email_html = "TNEB scraping has failed. PFA Logs & Screenshot for the code run."
                email_mime = self.create_mime_type(email_html, attachment=zip_buffer, attachment_type='zip', failureFlag=True)
            
            elif len(bill_details) == 0:

                self.logger.info("No scraped results were found.")
                
                # Attach image to email for confirmation
                with open(f"{self.debug_path}/bill-details.png", "rb") as f:
                    img_buffer = f.read()

                email_html = TNEBMailer.dataframe_to_html(bill_details)
                email_mime = self.create_mime_type(email_html, attachment=img_buffer, attachment_type='img', failureFlag=False)

            else:

                email_html = TNEBMailer.dataframe_to_html(bill_details)
                self.logger.info("Dataframe converted to HTML.")

                email_mime = self.create_mime_type(email_html)
                self.logger.info("Mimetype File created with HTML.")

            server.starttls(context=context)
            server.login(self.FROM_EMAIL, self.FROM_EMAIL_PWD)
            self.logger.info("Authentication GMAIL Server with provided credentials successful.")

            server.sendmail(
                self.FROM_EMAIL,
                self.TO_EMAIL,
                email_mime.as_string(),
            )

            self.logger.info ("Mail sent successfully.")

        except Exception as e:
            self.logger.info(f"An error occurred while sending email: {e}")

        finally:
            server.close()

    def dataframe_to_html(bill_details: pd.DataFrame) -> str:

        # Create the HTML Table
        html_table = """
            <html lang='en'>
                <head>
                    <style>
                        table {
                        border-collapse: collapse;
                        width: 100%;
                        }

                        th, td {
                        border: 1px solid #dddddd;
                        text-align: left;
                        padding: 8px;
                        }

                        th {
                        background-color: #f2f2f2;
                        }

                        tr:nth-child(even) {
                        background-color: #f9f9f9;
                        }
                    </style>
                </head>
                <body>
                    <table>
                    <thead>
                    <tr>
                        <th>S No.</th>
                        <th>Consumer Number</th>
                        <th>Portion Code</th>
                        <th>Bill Amount</th>
                        <th>Due Date</th>
                    </tr>
                    </thead>
                    <tbody>
        """

        if len(bill_details) > 0:
            for index, bill in bill_details.iterrows():
                html_table += f"""
                <tr>
                    <td>{index + 1}</td>
                    <td>{bill['Consumer No']}</td>
                    <td>{bill['Portion']}</td>
                    <td>{bill['Bill Amt (Rs)']}</td>
                    <td>{bill['Due Date']}</td>
                </tr>
                """

        else:
           html_table += """
            <tr>
                <td colspan='5'>No scraped results were found. All bills have been paid.</td>
            </tr>
            """ 

        html_table += "</tbody></table></body></html>"

        return html_table

    def create_mime_type(self, html_str, attachment=None, attachment_type=None, failureFlag=False):
        message = MIMEMultipart("alternative")
        message["Subject"] = f"Job {'failed' if failureFlag else 'succeeded'}: TamilNadu EB Bill - {dt.date.today().strftime('%d %B %Y')}"
        message["From"] = self.FROM_EMAIL
        message["To"] = self.TO_EMAIL

        html_part = MIMEText(html_str, "html")
        message.attach(html_part)

        # Attach the log file if provided
        if attachment:

            if attachment_type == 'zip':
                attachment_part = MIMEBase("application", "octet-stream")
                attachment_part.set_payload(attachment.getvalue())
                encoders.encode_base64(attachment_part)
                attachment_part.add_header("Content-Disposition", f"attachment; filename=logs-{int(dt.datetime.now().timestamp() * 1000000)}.zip")
                message.attach(attachment_part)

            elif attachment_type == 'img':
                attachment_part = MIMEImage(attachment, name="Bill-Details-Empty.png")
                message.attach(attachment_part)

            else:
                self.logger.error(f"Unsupported attachment type provided: {attachment_type}")
                raise NotImplementedError(f"{attachment_type} type of attachments are not suppoerted yet.")

        return message