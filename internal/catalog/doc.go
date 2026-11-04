// Package catalog loads the component catalog: the component manifests and
// template files under components/, read from an fs.FS, the contribution
// points kinds declare in kind.yaml, and the base layer in components/base/
// with its contribution points. It validates every manifest, enforces the
// file naming rules, and lists the candidate selections, that is, one
// component per required kind and at most one per optional kind with all
// required traits met.
package catalog
