package mutex

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES,
// INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
// WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE
// USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Extracted from https://h12.io/article/go-pattern-context-aware-lock

import (
	"context"
)

type CtxMutex struct {
	ch chan struct{}
}

func NewCtxMutex(maxConcurrent uint) *CtxMutex {
	return &CtxMutex{ch: make(chan struct{}, maxConcurrent)}
}

func (mu *CtxMutex) Lock(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case mu.ch <- struct{}{}:
		return true
	}
}

func (mu *CtxMutex) Unlock() {
	<-mu.ch
}

func (mu *CtxMutex) Locked() bool {
	return len(mu.ch) > 0
}
