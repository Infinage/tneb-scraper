import pandas as pd
import datetime as dt
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
import ssl, smtplib

class TNEBMailer:

    def __init__(self, FROM_EMAIL, FROM_EMAIL_PWD, TO_EMAIL):
        self.FROM_EMAIL = FROM_EMAIL
        self.TO_EMAIL = TO_EMAIL
        self.FROM_EMAIL_PWD = FROM_EMAIL_PWD

    def send_mail(self, bill_details: pd.DataFrame):

        server = smtplib.SMTP("smtp.gmail.com", 587)
        context = ssl.create_default_context()

        try:
            if len(bill_details) == 0:
                raise Exception("No scraped results were found.")

            email_html = TNEBMailer.dataframe_to_html(bill_details)
            print("Dataframe converted to HTML.")

            email_mime = self.create_mime_type(email_html)
            print("Mimetype File created with HTML.")

            server.starttls(context=context)
            server.login(self.FROM_EMAIL, self.FROM_EMAIL_PWD)
            print("Authentication GMAIL Server with provided credentials successful.")

            server.sendmail(
                self.FROM_EMAIL,
                self.TO_EMAIL,
                email_mime.as_string(),
            )
            print ("Mail sent successfully.")

        except Exception as e:
            print(f"An error occurred while sending email: {e}")

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

        html_table += "</tbody></table></body></html>"

        return html_table

    def create_mime_type(self, html_str):
        message = MIMEMultipart("alternative")
        message["Subject"] = "TamilNadu EB Bill - " + dt.date.today().strftime("%d %B %Y")
        message["From"] = self.FROM_EMAIL
        message["To"] = self.TO_EMAIL

        html_part = MIMEText(html_str, "html")
        message.attach(html_part)
        return message