Feature: ConfigMap Scoped Annotations
    @disabled
    Scenario: Copy a single key from a ConfigMap related to the Service resource to the binding secret
        Given CRD "databases.postgresql.baiju.dev" contains the annotation "service.binding/certificate: path={.status.data.dbConfiguration},objectType=ConfigMap,sourceKey=certificate"
        And Resource "cmsa-1-configmap" is created
            """
            apiVersion: v1
            kind: ConfigMap
            metadata:
                name: cmsa-1-configmap
            data:
                certificate: "certificate value"
            """
        And Resource "cmsa-2-service" is created
            """
            apiVersion: postgresql.baiju.dev/v1alpha1
            kind: Database
            metadata:
                name: cmsa-1-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbConfiguration: cmsa-1-configmap    # ConfigMap
            """
        When Resource "cmsa-1" is created
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: cmsa-1
            spec:
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: cmsa-1-service
            """
        Then Secret "cmsa-1" contains "CERTIFICATE" key with value "certificate value"

    @disabled
    Scenario: Copy all keys from the ConfigMap related to the Service resource into the binding secret
        Given CRD "databases.postgresql.baiju.dev" contains the annotation "service.binding: path={.status.data.dbConfiguration},objectType=ConfigMap,elementType=map"
        And Resource "cmsa-2-configmap" is created
            """
            apiVersion: v1
            kind: ConfigMap
            metadata:
                name: cmsa-2-configmap
            data:
                timeout: 30
                certificate: certificate value
            """
        And Resource "cmsa-2-service" is created
            """
            apiVersion: postgresql.baiju.dev/v1alpha1
            kind: Database
            metadata:
                name: cmsa-2-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                    dbConfiguration: cmsa-2-configmap    # ConfigMap
            """
        When Resource "cmsa-2" is created
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: cmsa-2
            spec:
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: cmsa-2-service
            """
        Then Secret "cmsa-2" contains "CERTIFICATE" key with value "certificate value"
        And Secret "cmsa-2" contains "TIMEOUT" key with value "30"
