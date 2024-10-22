#import <ApplicationServices/ApplicationServices.h>

static const CGKeyCode kVK_ANSI_A                = 0x00;
static const CGKeyCode kVK_ANSI_S                = 0x01;
static const CGKeyCode kVK_ANSI_D                = 0x02;
static const CGKeyCode kVK_ANSI_F                = 0x03;
static const CGKeyCode kVK_ANSI_H                = 0x04;
static const CGKeyCode kVK_ANSI_G                = 0x05;
static const CGKeyCode kVK_ANSI_Z                = 0x06;
static const CGKeyCode kVK_ANSI_X                = 0x07;
static const CGKeyCode kVK_ANSI_C                = 0x08;
static const CGKeyCode kVK_ANSI_V                = 0x09;
static const CGKeyCode kVK_ANSI_B                = 0x0B;
static const CGKeyCode kVK_ANSI_Q                = 0x0C;
static const CGKeyCode kVK_ANSI_W                = 0x0D;
static const CGKeyCode kVK_ANSI_E                = 0x0E;
static const CGKeyCode kVK_ANSI_R                = 0x0F;
static const CGKeyCode kVK_ANSI_Y                = 0x10;
static const CGKeyCode kVK_ANSI_T                = 0x11;
static const CGKeyCode kVK_ANSI_1                = 0x12;
static const CGKeyCode kVK_ANSI_2                = 0x13;
static const CGKeyCode kVK_ANSI_3                = 0x14;
static const CGKeyCode kVK_ANSI_4                = 0x15;
static const CGKeyCode kVK_ANSI_6                = 0x16;
static const CGKeyCode kVK_ANSI_5                = 0x17;
static const CGKeyCode kVK_ANSI_Equal            = 0x18;
static const CGKeyCode kVK_ANSI_9                = 0x19;
static const CGKeyCode kVK_ANSI_7                = 0x1A;
static const CGKeyCode kVK_ANSI_Minus            = 0x1B;
static const CGKeyCode kVK_ANSI_8                = 0x1C;
static const CGKeyCode kVK_ANSI_0                = 0x1D;
static const CGKeyCode kVK_ANSI_RightBracket     = 0x1E;
static const CGKeyCode kVK_ANSI_O                = 0x1F;
static const CGKeyCode kVK_ANSI_U                = 0x20;
static const CGKeyCode kVK_ANSI_LeftBracket      = 0x21;
static const CGKeyCode kVK_ANSI_I                = 0x22;
static const CGKeyCode kVK_ANSI_P                = 0x23;
static const CGKeyCode kVK_ANSI_L                = 0x25;
static const CGKeyCode kVK_ANSI_J                = 0x26;
static const CGKeyCode kVK_ANSI_Quote            = 0x27;
static const CGKeyCode kVK_ANSI_K                = 0x28;
static const CGKeyCode kVK_ANSI_Semicolon        = 0x29;
static const CGKeyCode kVK_ANSI_Backslash        = 0x2A;
static const CGKeyCode kVK_ANSI_Comma            = 0x2B;
static const CGKeyCode kVK_ANSI_Slash            = 0x2C;
static const CGKeyCode kVK_ANSI_N                = 0x2D;
static const CGKeyCode kVK_ANSI_M                = 0x2E;
static const CGKeyCode kVK_ANSI_Period           = 0x2F;
static const CGKeyCode kVK_ANSI_Grave            = 0x32;
static const CGKeyCode kVK_ANSI_KeypadDecimal    = 0x41;
static const CGKeyCode kVK_ANSI_KeypadMultiply   = 0x43;
static const CGKeyCode kVK_ANSI_KeypadPlus       = 0x45;
static const CGKeyCode kVK_ANSI_KeypadClear      = 0x47;
static const CGKeyCode kVK_ANSI_KeypadDivide     = 0x4B;
static const CGKeyCode kVK_ANSI_KeypadEnter      = 0x4C;
static const CGKeyCode kVK_ANSI_KeypadMinus      = 0x4E;
static const CGKeyCode kVK_ANSI_KeypadEquals     = 0x51;
static const CGKeyCode kVK_ANSI_Keypad0          = 0x52;
static const CGKeyCode kVK_ANSI_Keypad1          = 0x53;
static const CGKeyCode kVK_ANSI_Keypad2          = 0x54;
static const CGKeyCode kVK_ANSI_Keypad3          = 0x55;
static const CGKeyCode kVK_ANSI_Keypad4          = 0x56;
static const CGKeyCode kVK_ANSI_Keypad5          = 0x57;
static const CGKeyCode kVK_ANSI_Keypad6          = 0x58;
static const CGKeyCode kVK_ANSI_Keypad7          = 0x59;
static const CGKeyCode kVK_ANSI_Keypad8          = 0x5B;
static const CGKeyCode kVK_ANSI_Keypad9          = 0x5C;

// keycodes for keys that are independent of keyboard layout
static const CGKeyCode kVK_Return                = 0x24;
static const CGKeyCode kVK_Tab                   = 0x30;
static const CGKeyCode kVK_Space                 = 0x31;
static const CGKeyCode kVK_Delete                = 0x33;
static const CGKeyCode kVK_Escape                = 0x35;
static const CGKeyCode kVK_Command               = 0x37;
static const CGKeyCode kVK_Shift                 = 0x38;
static const CGKeyCode kVK_CapsLock              = 0x39;
static const CGKeyCode kVK_Option                = 0x3A;
static const CGKeyCode kVK_Control               = 0x3B;
static const CGKeyCode kVK_RightCommand          = 0x36; // Out of order;
static const CGKeyCode kVK_RightShift            = 0x3C;
static const CGKeyCode kVK_RightOption           = 0x3D;
static const CGKeyCode kVK_RightControl          = 0x3E;
static const CGKeyCode kVK_Function              = 0x3F;
static const CGKeyCode kVK_F17                   = 0x40;
static const CGKeyCode kVK_VolumeUp              = 0x48;
static const CGKeyCode kVK_VolumeDown            = 0x49;
static const CGKeyCode kVK_Mute                  = 0x4A;
static const CGKeyCode kVK_F18                   = 0x4F;
static const CGKeyCode kVK_F19                   = 0x50;
static const CGKeyCode kVK_F20                   = 0x5A;
static const CGKeyCode kVK_F5                    = 0x60;
static const CGKeyCode kVK_F6                    = 0x61;
static const CGKeyCode kVK_F7                    = 0x62;
static const CGKeyCode kVK_F3                    = 0x63;
static const CGKeyCode kVK_F8                    = 0x64;
static const CGKeyCode kVK_F9                    = 0x65;
static const CGKeyCode kVK_F11                   = 0x67;
static const CGKeyCode kVK_F13                   = 0x69;
static const CGKeyCode kVK_F16                   = 0x6A;
static const CGKeyCode kVK_F14                   = 0x6B;
static const CGKeyCode kVK_F10                   = 0x6D;
static const CGKeyCode kVK_F12                   = 0x6F;
static const CGKeyCode kVK_F15                   = 0x71;
static const CGKeyCode kVK_Help                  = 0x72;
static const CGKeyCode kVK_Home                  = 0x73;
static const CGKeyCode kVK_PageUp                = 0x74;
static const CGKeyCode kVK_ForwardDelete         = 0x75;
static const CGKeyCode kVK_F4                    = 0x76;
static const CGKeyCode kVK_End                   = 0x77;
static const CGKeyCode kVK_F2                    = 0x78;
static const CGKeyCode kVK_PageDown              = 0x79;
static const CGKeyCode kVK_F1                    = 0x7A;
static const CGKeyCode kVK_LeftArrow             = 0x7B;
static const CGKeyCode kVK_RightArrow            = 0x7C;
static const CGKeyCode kVK_DownArrow             = 0x7D;
static const CGKeyCode kVK_UpArrow               = 0x7E;

// ISO keyboards only
static const CGKeyCode kVK_ISO_Section           = 0x0A;

// JIS keyboards only
static const CGKeyCode kVK_JIS_Yen               = 0x5D;
static const CGKeyCode kVK_JIS_Underscore        = 0x5E;
static const CGKeyCode kVK_JIS_KeypadComma       = 0x5F;
static const CGKeyCode kVK_JIS_Eisu              = 0x66;
static const CGKeyCode kVK_JIS_Kana              = 0x68;
