// Displays local macOS password, file picker and recovery dialogs through Standard Additions.
// Request and response travel on private pipes; secret values stay outside process arguments.
ObjC.import('AppKit')
const app = Application.currentApplication()
app.includeStandardAdditions = true
function run() {
    try {
        const input = $.NSFileHandle.fileHandleWithStandardInput.readDataToEndOfFile
        const req = JSON.parse(ObjC.unwrap($.NSString.alloc.initWithDataEncoding(input, $.NSUTF8StringEncoding)))
        let value
        if (req.Op === 'file') {
            const path = app.chooseFile({
                withPrompt:req.Title
            }).toString()
            const source = $.NSFileHandle.fileHandleForReadingAtPath(path)
            if (!source) return 'null'
            const data = source.readDataOfLength(16385)
            source.closeFile
            if (!data || Number(data.length) > 16384) return 'null'
            value = ObjC.unwrap(data.base64EncodedStringWithOptions(0))
        } else if (req.Op === 'save') {
            const path = app.chooseFileName({
                withPrompt:req.Title,defaultName:'apparat-recovery.json'
            }).toString()
            const data = $.NSData.alloc.initWithBase64EncodedStringOptions(req.Value,0)
            if (!data.writeToFileOptionsError(path,$.NSDataWritingWithoutOverwriting,null)) return 'null'
            value = ''
        } else if (req.Op === 'password' || req.Op === 'text') {
            value = app.displayDialog(req.Message,{
                withTitle:req.Title, defaultAnswer:'', hiddenAnswer:req.Op === 'password', buttons:['Cancel','Continue'],defaultButton:'Continue',cancelButton:'Cancel'
            }).textReturned
        } else if (req.Op === 'trust') {
            const alert = $.NSAlert.alloc.init
            alert.messageText = req.Title
            alert.informativeText = req.Message
            for (const option of req.Options) alert.addButtonWithTitle(option)
            alert.layout
            alert.window.setFrameDisplay($.NSScreen.mainScreen.frame, true)
            const result = alert.runModal - 1000
            if (result < 0 || result >= req.Options.length) return 'null'
            value = req.Options[result]
        } else if (req.Op === 'choose') {
            const selection = app.chooseFromList(req.Options,{
                withTitle:req.Title,withPrompt:req.Message
            })
            if (!selection) return 'null'
            value = selection[0]
        } else if (req.Op === 'notice') {
            app.displayDialog(req.Message,{
                withTitle:req.Title,buttons:['OK']
            })
            value = ''
        } else return 'null'
        return JSON.stringify(value)
    } catch (_) {
        return 'null'
    }
}
