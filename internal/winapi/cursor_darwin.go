//go:build darwin

package winapi

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation
#include <CoreGraphics/CoreGraphics.h>

static void moveMouse(double x, double y) {
    CGEventRef move = CGEventCreateMouseEvent(NULL, kCGEventMouseMoved,
        CGPointMake(x, y), kCGMouseButtonLeft);
    CGEventPost(kCGHIDEventTap, move);
    CFRelease(move);
}

static void getCursorPosition(double *x, double *y) {
    CGEventRef event = CGEventCreate(NULL);
    CGPoint cursor = CGEventGetLocation(event);
    *x = cursor.x;
    *y = cursor.y;
    CFRelease(event);
}

static void getMainDisplaySize(double *w, double *h) {
    CGDirectDisplayID mainDisplay = CGMainDisplayID();
    *w = (double)CGDisplayPixelsWide(mainDisplay);
    *h = (double)CGDisplayPixelsHigh(mainDisplay);
}

// Virtual screen bounds across all displays.
static void getVirtualBounds(double *x, double *y, double *w, double *h) {
    // CGDisplayBounds for the union of all active displays.
    uint32_t count = 0;
    CGGetActiveDisplayList(0, NULL, &count);
    if (count == 0) {
        *x = 0; *y = 0; *w = 1920; *h = 1080;
        return;
    }
    CGDirectDisplayID *displays = malloc(count * sizeof(CGDirectDisplayID));
    CGGetActiveDisplayList(count, displays, &count);

    double minX = 1e9, minY = 1e9, maxX = -1e9, maxY = -1e9;
    for (uint32_t i = 0; i < count; i++) {
        CGRect bounds = CGDisplayBounds(displays[i]);
        if (bounds.origin.x < minX) minX = bounds.origin.x;
        if (bounds.origin.y < minY) minY = bounds.origin.y;
        if (bounds.origin.x + bounds.size.width > maxX) maxX = bounds.origin.x + bounds.size.width;
        if (bounds.origin.y + bounds.size.height > maxY) maxY = bounds.origin.y + bounds.size.height;
    }
    free(displays);
    *x = minX; *y = minY;
    *w = maxX - minX; *h = maxY - minY;
}
*/
import "C"

import (
	"github.com/jarvis/mouse-mover/internal/engine"
)

type darwinController struct{}

// NewCursorController returns a macOS CursorController using Core Graphics.
func NewCursorController() engine.CursorController {
	return &darwinController{}
}

func (c *darwinController) GetPos() (engine.Point, error) {
	var x, y C.double
	C.getCursorPosition(&x, &y)
	return engine.Point{X: int32(x), Y: int32(y)}, nil
}

func (c *darwinController) MoveTo(p engine.Point) error {
	C.moveMouse(C.double(p.X), C.double(p.Y))
	return nil
}

func (c *darwinController) VirtualScreen() (engine.Rect, error) {
	var x, y, w, h C.double
	C.getVirtualBounds(&x, &y, &w, &h)
	return engine.Rect{
		Left:   int32(x),
		Top:    int32(y),
		Right:  int32(x) + int32(w),
		Bottom: int32(y) + int32(h),
	}, nil
}
