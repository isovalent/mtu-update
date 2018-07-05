Usage
-----

.. code-block:: shell-session

    $ ./mtu-update -h
    Update the MTU inside network namespaces.

    Usage:
      mtu-update [flags]

    Flags:
      -h, --help                  help for mtu-update
      -m, --mtu int               Base MTU to configure on links (0 for autodetect) (default 1500)
      -t, --tunnel-overhead int   Expected tunnel overhead for overlay traffic (default 50)
      -v, --verbose               Print verbose debug log messages

Update the MTU across a k8s cluster:

.. code-block:: shell-session

    $ kubectl create -f https://raw.githubusercontent.com/cilium/mtu-update/1.1/mtu-update.yaml
    $ kubectl get ds mtu-update -n kube-system
    NAME         DESIRED   CURRENT   READY     UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
    mtu-update   1         1         1         1            1           <none>          18s
    (Try again until all pods are ready)
    $ kubectl delete -f https://raw.githubusercontent.com/cilium/mtu-update/1.1/mtu-update.yaml

Contact
-------

If you have any questions feel free to contact us on `Slack <https://cilium.herokuapp.com/>`_.


License
-------

This project is licensed under the `Apache License, Version 2.0 <LICENSE>`_.
