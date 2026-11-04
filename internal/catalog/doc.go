// Package catalog loads the component catalog: the component manifests and
// template files under components/, read from an fs.FS, and the base layer
// in components/base/ with its contribution points. It validates every
// manifest, enforces the file naming rules, and lists the supported
// selections, that is, one component per kind with all required traits met.
package catalog
