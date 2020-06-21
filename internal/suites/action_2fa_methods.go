package suites

import (
	"context"
	"fmt"
	"testing"
)

func (wds *WebDriverSession) doChangeMethod(ctx context.Context, t *testing.T, method string) {
	wds.WaitElementLocatedByID(ctx, t, "methods-button").Click() //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	wds.WaitElementLocatedByID(ctx, t, "methods-dialog")
	wds.WaitElementLocatedByID(ctx, t, fmt.Sprintf("%s-option", method)).Click() //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
}
