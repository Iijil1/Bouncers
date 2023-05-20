# Bouncers

A bouncer is a turing machine M for which we can find values that make the following description work:

 Let S be a tm state, a, b, c, d, e, k integers >= 0 with c > 0, k >= 3 and k even, and w_0, w_1, ..., w_k words of tm symbols.
 Define the configuration `C(n) = { tape: 0^∞ w_0 w_1^n w_2 w_3^n w_4 ... w_k-1^n w_k 0^∞, state: S, pos d + e*n }`
 M reaches C(n) after `a + b*n + c*n^2` steps.

We call the words w_i with even i walls, as the machine bounces back and forth between these walls during simulation.
The words w_i with odd i are called repeaters, as the are repeated more and more often in the later configurations.

## Notation

I'll write configurations with the state between the words on the tape and an indicated read direction. So `w_i <S` means the head is on the last symbol of w_i, `S> w_i` means the head is on the first symbol of w_i.

I'll omit the 0^∞ at the beginning and end of the tape. w_0 and w_k will include leading and trailing 0s exactly if those were written by the tm.

# Proving Bouncers

If someone hands us the values that show a tm M is a bouncer we can verify that it reaches C(0) by direct simulation. Then we need to show that from C(n) it reaches C(n+1) in b+ c\*n steps to complete a proof by induction.

To make the induction step easier we put additional restrictions on the values:

 We only consider configurations with a fixed position: `C(n) = w_0 S> w_1^n w_2 ... w_k-1^n w_k`. Most bouncers will naturally reach this position at some point during their growth cycle. The few that don't are "unilateral translated bouncers to the right" in the bbchallenge.org zoology. For those we can swap all L and R transitions to get an equivalent tm that grows to the left and prove that instead.

 We split off a buffer word from w_0. Treating the buffer as seperate from w_0 and more as part of the tm state makes the proof easier. You can think of it like the symbols behind the state of a backward-symbol macro machine. So our configurations look like this: `C(n) = w_0 buf S> w_1^n w_2 ... w_k-1^n w_k`

## Transition rules

To go through the induction step we now need some transition rules. Specifically we need to know what M does on the seperate words w_i that it encounters during a growth cycle. Each rule starts with `buf S> word` or `word <S buf` and tells us how many steps it takes to get to `word' buf' S'>` or `<S' buf' word'`.

 During such a rule the tm is not allowed to leave the part of the tape that contains the word and buffer, except on the last step of the rule were it has to step outside.
 This puts it on the first or last symbol of the neighboring word and into the starting position for the next rule.

 For this purpose w_0 and w_k are considered to contain the infinite 0s at the tape ends, so if the simulator reaches those ends of the simulated tape extend it with 0.

 The very last rule is an exception and is allowed to end without stepping out of the word and buffer tape area. It has the form `word <S buf --> word' buf' S'> stub`. This exception is made to allow the leftmost wall to move to the left over time.

The rules are not allowed to change the length of the buffer, so `len(buf) = len(buf')` is required.

For each individual rule we can check that it works by direct simulation.

## Chain rules

To be able to apply a transition rule to a repeater it needs to be of the form `buf S> word --> word' buf S>` or `word <S buf --> <S buf word'`. In particular the buffer, state and direction need to be the same at the start and end of the rule. Then it can be shown by induction that `buf S> word^n --> word'^n buf S>` or `word^n <S buf --> <S buf word'^n` by chaining the rule application n times.

Chain rules are not allowed to have `len(word) = 0`, as an empty repeater can be avoided by just merging the walls on either side.

## Applying Rules

To verify a bouncer we require a list of rules r_0, r_1, ... r_l with l odd where r_i is a chain rule for all even i. We can then successively apply them to the configuration C(n). With each application we need to change the word at the current position and then update the state, buffer, direction and position according to the rule.
 
 For example applying a rule could turn `x y^n buf1 B> u` into `x y^n <C buf2 v`.

 (Every rule application changes the position +-1, every second word is a repeater and every second transition rule is a chain rule. So only chain rules will be applied to repeaters and we can just make the changes to the entire w_i^n block at once as if it was a single application of the rule.)

At each step we of course need to verify that the start conditions of the rule are met by checking the word at the current positon, the state, buffer and direction.

After applying all the given rules we know that `C(n) = w_0 buf S> w_1^n w_2 ... w_k-1^n w_k --> C'(n) = w'_0 buf' S'> stub w'_1^n w'_2 ... w'_k-1^n w'_k`

## Finish the induction

We now need to show that `C'(n) = C(n+1)`. w_0, buf, S, position and direction are just checked for equality with their counterparts. The interesting stuff happens to the right of the head where there was actual growth from the repeaters.
 For C(n+1) we split each repeaters w_i^n+1 up into w_i w_i^n. We then add w_i to the wall to the left of it. We find that the tape to the right of the head looks like this: `w_1 w_1^n (w_2 w_3) w_3^n ... (w_k-2 w_k-1) w_k-1^n w_k`. We now use the fact that `(xy)^n x = x (yx)^n` to align all the repeaters in the representation as far to the right as possible. We get new words v_1, ... v_k and know that the right part of the C(n+1) tape looks like `v_0 v_1^n v_2 ... v_k-1^n v_k`.

 For C'(n) we start with `stub w'_1^n w'_2 ... w'_k-1^n w'_k` and also align all the repeaters as far to the right as possible. We get `v'_0 v'_1^n v'_2 ... v'_k-1^n v'_k`.

We can now simply compare all v_i to the corresponding v'_i. If all are equal the induction is finished and we have a bouncer.

# Full Certificates

With -pm=2  I print the full certificate for bouncers I find in JSON.

This includes the tm in standard text format, whether it needs to be mirrored for the proof, the information about C(n), the number of steps until the tm reaches C(0) and a full list of the necessary rules.

For each rule I give the start conditions, the end conditions, the number of steps it takes and whether it takes place at the end of the tape.

# Short Certificates

It is possible to derive the rules from C(n) with just a little bit of extra information: For most rules we just simulate the tm until we run out of the allowed tape segment. This gives us the end conditions for this rule and the start conditions for the next. Only the last rule can stop early. If we know the sum of steps taken in all rules we can keep track of the steps we are still allowed to use and know when this early stop happens.

So with -pm=1 I print the short certificate that is like the full certificate, but with the number of steps taken across all rules instead of the full rule list.

# Finding Bouncers

After simulating the tm for a number of steps we check whether any record breaking configurations are in the quadratic time grwoth sequence required by bouncers. If we find such records we try to split the corresponding tapes in walls and repeaters. If successful we can use that like a short certificate to derive the rules and prove that the tm is a bouncer.

The problem is splitting up the tape. The approach taken in this decider is to color each symbol on the tapes with a lot of context. The color is determined by the sequence of states the tm was in when it visited each cell in an area around the symbol since the last record.

After coloring the tape it is split into walls and repeaters greedily by comparing 2 successive records.

The buffersize is determined by looking at the sequence of steps between turn arounds between left and right movements taken to get from each record to the next. We smooth those sequences by combining movements to what the would have been with a bigger buffer until each term includes movement through a repeater. At that point the buffer is big enough that the tm never turns around outside the buffer during a chain rule, and only turns around at most once outside the buffer for the wall transitions.