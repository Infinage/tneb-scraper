import * as Tesseract from "tesseract.js";

export const extractText = async (fp: string): Promise<string> => {
    const worker = await Tesseract.createWorker();
    await worker.loadLanguage("eng");
    await worker.initialize("eng", Tesseract.OEM.LSTM_ONLY);
    await worker.setParameters({
            tessedit_char_whitelist: '0123456789',
            tessedit_pageseg_mode: Tesseract.PSM.SINGLE_WORD,
    });
    const { data: {text} } = await worker.recognize(fp);
    await worker.terminate();
    return text;
};