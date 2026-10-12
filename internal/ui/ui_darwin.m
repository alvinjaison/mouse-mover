#import <Cocoa/Cocoa.h>

// Go callbacks
extern void goStartEngine(int durationMin, int intervalSec, int patternIdx, int delta);
extern void goStopEngine(void);
extern int  goIsRunning(void);
extern void goSaveConfig(int durationMin, int intervalSec, int patternIdx, int delta);

// ---- App Delegate (menu bar app) ----

@interface MMAppDelegate : NSObject <NSApplicationDelegate, NSMenuDelegate>
@property (strong) NSStatusItem *statusItem;
@property (strong) NSMenu *statusMenu;
@property (strong) NSTimer *updateTimer;

// Config values (set from Go before launch)
@property int cfgDurationMin;
@property int cfgIntervalSec;
@property int cfgPatternIdx;
@property int cfgDelta;
@property int cfgAutoStart;

// Menu items that need state updates
@property (strong) NSMenuItem *startItem;
@property (strong) NSMenuItem *stopItem;
@property (strong) NSMenuItem *statusMenuItem;

// Duration sub-items
@property (strong) NSMenuItem *dur30m;
@property (strong) NSMenuItem *dur1h;
@property (strong) NSMenuItem *dur2h;
@property (strong) NSMenuItem *dur4h;
@property (strong) NSMenuItem *dur8h;
@property (strong) NSMenuItem *durOff;

// Interval sub-items
@property (strong) NSMenuItem *int10s;
@property (strong) NSMenuItem *int30s;
@property (strong) NSMenuItem *int1m;
@property (strong) NSMenuItem *int2m;
@property (strong) NSMenuItem *int5m;
@end

@implementation MMAppDelegate

- (void)applicationDidFinishLaunching:(NSNotification *)notification {
    // Create status bar item
    self.statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
    self.statusItem.button.title = @"🖱 Stopped";
    self.statusItem.button.toolTip = @"Mouse Mover";

    // Build menu
    self.statusMenu = [[NSMenu alloc] initWithTitle:@"Mouse Mover"];
    self.statusItem.menu = self.statusMenu;

    // Status display (disabled, just for info)
    self.statusMenuItem = [[NSMenuItem alloc] initWithTitle:@"Status: Stopped" action:nil keyEquivalent:@""];
    [self.statusMenuItem setEnabled:NO];
    [self.statusMenu addItem:self.statusMenuItem];
    [self.statusMenu addItem:[NSMenuItem separatorItem]];

    // Start / Stop
    self.startItem = [[NSMenuItem alloc] initWithTitle:@"Start" action:@selector(onStart:) keyEquivalent:@"s"];
    self.startItem.target = self;
    [self.statusMenu addItem:self.startItem];

    self.stopItem = [[NSMenuItem alloc] initWithTitle:@"Stop" action:@selector(onStop:) keyEquivalent:@"x"];
    self.stopItem.target = self;
    self.stopItem.enabled = NO;
    [self.statusMenu addItem:self.stopItem];
    [self.statusMenu addItem:[NSMenuItem separatorItem]];

    // ---- Duration submenu ----
    NSMenuItem *durParent = [[NSMenuItem alloc] initWithTitle:@"Run Duration" action:nil keyEquivalent:@""];
    NSMenu *durMenu = [[NSMenu alloc] initWithTitle:@"Run Duration"];

    self.durOff = [[NSMenuItem alloc] initWithTitle:@"Indefinite" action:@selector(onDuration:) keyEquivalent:@""];
    self.durOff.target = self; self.durOff.tag = 0;
    self.dur30m = [[NSMenuItem alloc] initWithTitle:@"30 minutes" action:@selector(onDuration:) keyEquivalent:@""];
    self.dur30m.target = self; self.dur30m.tag = 30;
    self.dur1h = [[NSMenuItem alloc] initWithTitle:@"1 hour" action:@selector(onDuration:) keyEquivalent:@""];
    self.dur1h.target = self; self.dur1h.tag = 60;
    self.dur2h = [[NSMenuItem alloc] initWithTitle:@"2 hours" action:@selector(onDuration:) keyEquivalent:@""];
    self.dur2h.target = self; self.dur2h.tag = 120;
    self.dur4h = [[NSMenuItem alloc] initWithTitle:@"4 hours" action:@selector(onDuration:) keyEquivalent:@""];
    self.dur4h.target = self; self.dur4h.tag = 240;
    self.dur8h = [[NSMenuItem alloc] initWithTitle:@"8 hours" action:@selector(onDuration:) keyEquivalent:@""];
    self.dur8h.target = self; self.dur8h.tag = 480;

    [durMenu addItem:self.durOff];
    [durMenu addItem:self.dur30m];
    [durMenu addItem:self.dur1h];
    [durMenu addItem:self.dur2h];
    [durMenu addItem:self.dur4h];
    [durMenu addItem:self.dur8h];

    durParent.submenu = durMenu;
    [self.statusMenu addItem:durParent];
    [self updateDurationChecks];

    // ---- Interval submenu ----
    NSMenuItem *intParent = [[NSMenuItem alloc] initWithTitle:@"Move Interval" action:nil keyEquivalent:@""];
    NSMenu *intMenu = [[NSMenu alloc] initWithTitle:@"Move Interval"];

    self.int10s = [[NSMenuItem alloc] initWithTitle:@"10 seconds" action:@selector(onInterval:) keyEquivalent:@""];
    self.int10s.target = self; self.int10s.tag = 10;
    self.int30s = [[NSMenuItem alloc] initWithTitle:@"30 seconds" action:@selector(onInterval:) keyEquivalent:@""];
    self.int30s.target = self; self.int30s.tag = 30;
    self.int1m = [[NSMenuItem alloc] initWithTitle:@"1 minute" action:@selector(onInterval:) keyEquivalent:@""];
    self.int1m.target = self; self.int1m.tag = 60;
    self.int2m = [[NSMenuItem alloc] initWithTitle:@"2 minutes" action:@selector(onInterval:) keyEquivalent:@""];
    self.int2m.target = self; self.int2m.tag = 120;
    self.int5m = [[NSMenuItem alloc] initWithTitle:@"5 minutes" action:@selector(onInterval:) keyEquivalent:@""];
    self.int5m.target = self; self.int5m.tag = 300;

    [intMenu addItem:self.int10s];
    [intMenu addItem:self.int30s];
    [intMenu addItem:self.int1m];
    [intMenu addItem:self.int2m];
    [intMenu addItem:self.int5m];

    intParent.submenu = intMenu;
    [self.statusMenu addItem:intParent];
    [self updateIntervalChecks];

    // ---- Pattern submenu ----
    NSMenuItem *patParent = [[NSMenuItem alloc] initWithTitle:@"Pattern" action:nil keyEquivalent:@""];
    NSMenu *patMenu = [[NSMenu alloc] initWithTitle:@"Pattern"];
    NSArray *patNames = @[@"Nudge", @"Jiggle", @"Square"];
    for (int i = 0; i < 3; i++) {
        NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:patNames[i] action:@selector(onPattern:) keyEquivalent:@""];
        item.target = self;
        item.tag = i;
        if (i == self.cfgPatternIdx) item.state = NSControlStateValueOn;
        [patMenu addItem:item];
    }
    patParent.submenu = patMenu;
    [self.statusMenu addItem:patParent];

    [self.statusMenu addItem:[NSMenuItem separatorItem]];

    // Quit
    NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"Quit" action:@selector(onQuit:) keyEquivalent:@"q"];
    quitItem.target = self;
    [self.statusMenu addItem:quitItem];

    // Status update timer (1s)
    self.updateTimer = [NSTimer scheduledTimerWithTimeInterval:1.0
        target:self selector:@selector(updateUI) userInfo:nil repeats:YES];

    // Auto-start if configured
    if (self.cfgAutoStart) {
        [self onStart:nil];
    }
}

- (void)onStart:(id)sender {
    goSaveConfig(self.cfgDurationMin, self.cfgIntervalSec, self.cfgPatternIdx, self.cfgDelta);
    goStartEngine(self.cfgDurationMin, self.cfgIntervalSec, self.cfgPatternIdx, self.cfgDelta);
    [self updateUI];
}

- (void)onStop:(id)sender {
    goStopEngine();
    [self updateUI];
}

- (void)onDuration:(NSMenuItem *)sender {
    self.cfgDurationMin = (int)sender.tag;
    [self updateDurationChecks];
    goSaveConfig(self.cfgDurationMin, self.cfgIntervalSec, self.cfgPatternIdx, self.cfgDelta);
}

- (void)onInterval:(NSMenuItem *)sender {
    self.cfgIntervalSec = (int)sender.tag;
    [self updateIntervalChecks];
    goSaveConfig(self.cfgDurationMin, self.cfgIntervalSec, self.cfgPatternIdx, self.cfgDelta);
}

- (void)onPattern:(NSMenuItem *)sender {
    self.cfgPatternIdx = (int)sender.tag;
    // Update checkmarks
    NSMenu *menu = sender.menu;
    for (NSMenuItem *item in menu.itemArray) {
        item.state = (item.tag == self.cfgPatternIdx) ? NSControlStateValueOn : NSControlStateValueOff;
    }
    goSaveConfig(self.cfgDurationMin, self.cfgIntervalSec, self.cfgPatternIdx, self.cfgDelta);
}

- (void)onQuit:(id)sender {
    goStopEngine();
    [NSApp terminate:nil];
}

- (void)updateUI {
    int running = goIsRunning();
    if (running) {
        self.statusItem.button.title = @"🖱 Running";
        self.statusMenuItem.title = @"Status: Running";
        self.startItem.enabled = NO;
        self.stopItem.enabled = YES;
    } else {
        self.statusItem.button.title = @"🖱 Stopped";
        self.statusMenuItem.title = @"Status: Stopped";
        self.startItem.enabled = YES;
        self.stopItem.enabled = NO;
    }
}

- (void)updateDurationChecks {
    NSArray<NSMenuItem*> *items = @[self.durOff, self.dur30m, self.dur1h, self.dur2h, self.dur4h, self.dur8h];
    for (NSMenuItem *item in items) {
        item.state = (item.tag == self.cfgDurationMin) ? NSControlStateValueOn : NSControlStateValueOff;
    }
}

- (void)updateIntervalChecks {
    NSArray<NSMenuItem*> *items = @[self.int10s, self.int30s, self.int1m, self.int2m, self.int5m];
    for (NSMenuItem *item in items) {
        item.state = (item.tag == self.cfgIntervalSec) ? NSControlStateValueOn : NSControlStateValueOff;
    }
}

- (void)applicationWillTerminate:(NSNotification *)notification {
    goStopEngine();
}
@end

// ---- Entry point called from Go ----

void RunCocoaApp(int durationMinutes, int moveIntervalSeconds, int patternIndex,
                 int deltaPixels, int startMinimized, int autoStart) {
    @autoreleasepool {
        NSApplication *app = [NSApplication sharedApplication];
        [app setActivationPolicy:NSApplicationActivationPolicyAccessory]; // menu bar only, no dock icon

        MMAppDelegate *delegate = [[MMAppDelegate alloc] init];
        delegate.cfgDurationMin = durationMinutes;
        delegate.cfgIntervalSec = moveIntervalSeconds;
        delegate.cfgPatternIdx = patternIndex;
        delegate.cfgDelta = deltaPixels;
        delegate.cfgAutoStart = autoStart;

        app.delegate = delegate;
        [app run]; // blocks until quit
    }
}

void CocoaStop(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp terminate:nil];
    });
}
