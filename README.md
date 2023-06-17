## Getting started

1. Clone the github repository
2. Run `npm install` to install all the project dependencies
3. The repository npm scripts assumes that you have the following cmd tools installed: `rm`, `mkdir`, `tsc`, `cp`, `npm`, `cd`, `node`, `pkg`
4. We have two files that must be manually created: `eb-mapping.ts` & `.env` in the root directory

```
// Sample eb-mapping.ts
const ebMapping: Map<string, string> =  new Map([
    ["08401102339", "A"],
    ["08401102349", "B"],
    ["08402302339", "C"],
])

export default ebMapping;
```

```
// Sample _.env_ file:
GMAIL_APP_PWD=
GMAIL_FROM_ADDRESS=
GMAIL_TO_ADDRESS=
TNEB_LOGIN_URL=
TNEB_PASSWORD=
TNEB_USERNAME=
RETRY_ATTEMPTS=3
CRON_EXP="0 0 8 10 * *"
CRON_EXP_DESC="10th of each month by 8:00 AM"
```
5. Taking the project for a test run: `npm run dev:start`
6. Build the project to an executable: `npm run package`. Creates an executable in the _build/_ directory