# Instance Actor

This **Instance Actor** is exposed at the  `/actor` end-point.

Is an **ActivityPub** **Actor** of type `Application`.

This actor exists (as part of **tempfed**) to point **ActivityPub** **Actors** on other servers to the **shared-inbox** (exposed at the `/inbox` end-point), so that they (**ActivityPub** **Actors** on other servers) will send their _posts_ there.
