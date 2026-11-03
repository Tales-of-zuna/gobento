// Package catalog loads the component catalog: the component manifests and
// template files under components/, read from an fs.FS. It validates every
// manifest, enforces the file naming rules, and lists the supported
// selections, that is, one component per kind with all required traits met.
package catalog
