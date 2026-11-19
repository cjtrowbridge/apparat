## TODO

This file is an asynchronous inbox for user-captured tasks and questions. It is not execution authority, a roadmap, or an approved plan. Agents should read it for context, but only act on items here when the user explicitly asks to pull one into the active plan.

Key: `[ ]` pending, `[-]` in progress, `[x]` complete

- [x] make the tab button toolbar default to the bottom on vertical screens
- [x] make the projects tab be the one that's open by default on launching the app, but keep the settings option to change that to whatever the user wants
- [x] opening the settings tab should automatically check for updates
- [x] scrolling speed is inconsistent across tabs. tabs with more content scroll much faster. it seems like the math is set so that moving your finger from the top to bottom of the screen goes from the absolute top to absolute bottom of the content, rather than havaing the content scroll exactly as much as you move your finger. this should work the other way. also, scrolling up on a screen like projects when it only has one button causes the top of the content to scroll down. this behavior is a regression. It should not be possible to scroll above the top of the content or below the bottom of the content.
- [x] the whisper selection menu should explain which one is currently selected/loaded and there needs to be a dialog/explanation of downloading another model and then that one should be shown as loaded
