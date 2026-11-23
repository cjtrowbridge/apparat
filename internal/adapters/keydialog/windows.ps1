# Displays private WinForms dialogs and protects secrets with current-user DPAPI.
# All values are read from standard input and returned as JSON without logging.
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Windows.Forms,System.Drawing,System.Security
[Console]::InputEncoding = [Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
$r = [Console]::In.ReadToEnd() | ConvertFrom-Json
function Prompt {
    switch ($r.Op) {
        'available' {
            return 'Use Windows sign-in'
        }
        'store' {
            return [Convert]::ToBase64String([Security.Cryptography.ProtectedData]::Protect([Text.Encoding]::UTF8.GetBytes($r.Value),$null,'CurrentUser'))
        }
        'load' {
            return [Text.Encoding]::UTF8.GetString([Security.Cryptography.ProtectedData]::Unprotect([Convert]::FromBase64String($r.Value),$null,'CurrentUser'))
        }
        'remove' {
            return ''
        }
        'file' {
            $d = New-Object Windows.Forms.OpenFileDialog
            if ($d.ShowDialog() -ne 'OK' -or (Get-Item -LiteralPath $d.FileName).Length -gt 16384) {
                return
            }
            return [Convert]::ToBase64String([IO.File]::ReadAllBytes($d.FileName))
        }
        'save' {
            $d = New-Object Windows.Forms.SaveFileDialog
            $d.FileName = 'apparat-recovery.json'
            if ($d.ShowDialog() -ne 'OK') {
                return
            }
            $f=[IO.File]::Open($d.FileName,'CreateNew')
            try {
                $b=[Convert]::FromBase64String($r.Value)
                $f.Write($b,0,$b.Length)
            }
            finally {
                $f.Dispose()
            }
            return ''
        }
    }
    $form = New-Object Windows.Forms.Form
    $form.Text = $r.Title
    $form.AutoSize = $true
    if ($r.Op -eq 'trust') {
        $form.FormBorderStyle = 'None'
        $form.WindowState = 'Maximized'
    }
    $form.StartPosition = 'CenterScreen'
    $panel = New-Object Windows.Forms.FlowLayoutPanel
    $panel.FlowDirection = 'TopDown'
    $panel.AutoSize = $true
    $panel.Padding=12
    $form.Controls.Add($panel)
    $label=New-Object Windows.Forms.Label
    $label.MaximumSize=New-Object Drawing.Size(450,0)
    $label.AutoSize=$true
    $label.Text=$r.Message
    $panel.Controls.Add($label)
    $choices=@($r.Options)
    if ($r.Op -in 'password','text') {
        $entry=New-Object Windows.Forms.TextBox
        $entry.UseSystemPasswordChar=($r.Op -eq 'password')
        $entry.Width=300
        $panel.Controls.Add($entry)
        $choices=@('Continue')
    }
    foreach($choice in $choices){
        if (!$choice) {
            continue
        }
        $b=New-Object Windows.Forms.Button
        $b.AutoSize=$true
        $b.Text=$choice
        $b.Add_Click( {
            $form.Tag=$this.Text
            $form.DialogResult='OK'
        }
        )
        $panel.Controls.Add($b)
    }
    $cancel=New-Object Windows.Forms.Button
    $cancel.Text='Close'
    $cancel.DialogResult='Cancel'
    $form.CancelButton=$cancel
    $panel.Controls.Add($cancel)
    if ($form.ShowDialog() -ne 'OK') {
        return
    }
    if ($r.Op -in 'password','text') {
        return $entry.Text
    }
    return $form.Tag
}
try {
    $answer=Prompt
    [Console]::Write((ConvertTo-Json -InputObject $answer -Compress))
}
catch {
    [Console]::Write('null')
}
