package middlewares

import (
	"errors"
	"net/http"
	"path"
	"sort"
	"sync"

	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/constant"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/app-devper/um-api/sessionclient/ginauth"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// A Policy is who may call a route. Every pos route is registered through
// one, so who may call it is read from its name, not pieced together from a
// chain of middleware; docs/route-policies.txt lists them all.
//
//   - Staff: an active employee, in their branch.
//   - BranchAdmin: Staff who is at least a UM Admin.
//   - SignedIn: any live session whose employee record, if any, is active.
//     For shop-wide reads that need no branch.
//   - ShopAdmin: SignedIn and at least a UM Admin. For shop-wide writes.
type Policy struct {
	name     string
	handlers []gin.HandlerFunc
}

// Policies are the four Policies a feature registers routes with.
type Policies struct {
	Staff, BranchAdmin, SignedIn, ShopAdmin Policy
}

// NewPolicies builds the Policies over UM's session check and pos's
// employee and branch records.
func NewPolicies(auth *ginauth.Auth, employees repositories.IEmployee, branches repositories.IBranch) Policies {
	session := RequireSession(auth)
	branch := RequireBranch(employees, branches)
	active := RequireActiveEmployee(employees)
	admin := auth.AtLeast(sessionclient.RoleAdmin)
	return Policies{
		Staff:       Policy{"staff", []gin.HandlerFunc{session, branch}},
		BranchAdmin: Policy{"branch-admin", []gin.HandlerFunc{session, branch, admin}},
		SignedIn:    Policy{"signed-in", []gin.HandlerFunc{session, active}},
		ShopAdmin:   Policy{"shop-admin", []gin.HandlerFunc{session, active, admin}},
	}
}

// On registers routes on group under this Policy.
func (p Policy) On(group *gin.RouterGroup) Routes { return Routes{group, p} }

// Routes registers routes on one group under one Policy.
type Routes struct {
	group  *gin.RouterGroup
	policy Policy
}

func (r Routes) GET(relativePath string, h ...gin.HandlerFunc) {
	r.handle(http.MethodGet, relativePath, h)
}
func (r Routes) POST(relativePath string, h ...gin.HandlerFunc) {
	r.handle(http.MethodPost, relativePath, h)
}
func (r Routes) PUT(relativePath string, h ...gin.HandlerFunc) {
	r.handle(http.MethodPut, relativePath, h)
}
func (r Routes) PATCH(relativePath string, h ...gin.HandlerFunc) {
	r.handle(http.MethodPatch, relativePath, h)
}
func (r Routes) DELETE(relativePath string, h ...gin.HandlerFunc) {
	r.handle(http.MethodDelete, relativePath, h)
}

func (r Routes) handle(method, relativePath string, h []gin.HandlerFunc) {
	chain := append(append([]gin.HandlerFunc{}, r.policy.handlers...), h...)
	r.group.Handle(method, relativePath, chain...)
	full := path.Join(r.group.BasePath(), relativePath)
	registered.Lock()
	registered.byRoute[method+" "+full] = r.policy.name
	registered.Unlock()
}

var registered = struct {
	sync.Mutex
	byRoute map[string]string
}{byRoute: map[string]string{}}

// RoutePolicies lists every route registered through a Policy as
// "METHOD PATH POLICY", sorted.
func RoutePolicies() []string {
	registered.Lock()
	defer registered.Unlock()
	lines := make([]string, 0, len(registered.byRoute))
	for route, policy := range registered.byRoute {
		lines = append(lines, route+" "+policy)
	}
	sort.Strings(lines)
	return lines
}

// RequireActiveEmployee refuses a caller whose employee record is not
// active. A caller with no employee record passes: shop-wide routes do not
// need one, and the first Admin sets the shop up before any exist.
func RequireActiveEmployee(employees repositories.IEmployee) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		employee, err := employees.GetEmployeeByUserId(ctx.GetString("UserId"))
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
		case err != nil:
			errcode.Abort(ctx, http.StatusForbidden, errcode.AU_FORBIDDEN_001, "employee lookup failed")
			return
		case employee.Status != "" && employee.Status != constant.ACTIVE:
			errcode.Abort(ctx, http.StatusForbidden, errcode.AU_FORBIDDEN_001, "employee inactive")
			return
		}
		ctx.Next()
	}
}
