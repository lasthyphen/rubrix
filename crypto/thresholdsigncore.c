
#include "thresholdsign_include.h"
#include "dkg_include.h"

// Computes the Lagrange coefficient L(i+1) at 0 with regards to the range [signers(0)+1..signers(t)+1]
// and stores it in res, where t is the degree of the polynomial P
static void Zr_lagrangeCoefficientAtZero(bn_st* res, const int i, const uint32_t* signers, const int len){
    // r is the order of G1 and G2
    bn_st r, r_2;
    bn_new(&r);
    g2_get_ord(&r);
    // (r-2) is needed to compute the inverse in Zr
    // using little Fermat theorem
    bn_new(&r_2);
    bn_sub_dig(&r_2, &r, 2);
    // Barett reduction constant
    // TODO: hardcode u
    bn_st u;
    bn_new(&u)
    bn_mod_pre_barrt(&u, &r);

    // temp buffers
    bn_st acc, inv, base, numerator;
    bn_new(&inv);
    bn_new(&base);
    bn_new_size(&base, BITS_TO_DIGITS(Fr_BITS))
    bn_new(&acc);
    bn_new(&numerator);
    bn_new_size(&acc, BITS_TO_DIGITS(3*Fr_BITS));

    // the accumulator of the largarnge coeffiecient 
    // the sign (sign of acc) is equal to 1 if acc is positive, 0 otherwise
    bn_set_dig(&acc, 1);
    int sign = 1;

    // loops is the maximum number of loops that takes the accumulator to 
    // overflow modulo r. Mainly the highest k such that fact(MAX_IND)/fact(MAX_IND-k) < r
    const int loops = MAX_IND_LOOPS;
    int k,j = 0;
    while (j<len) {
        bn_set_dig(&base, 1);
        bn_set_dig(&numerator, 1);
        for (k = j; j < MIN(len, k+loops); j++){
            if (signers[j]==i) 
                continue;
            if (signers[j]<i) 
                sign ^= 1;
            bn_mul_dig(&base, &base, abs((int)signers[j]-i));
            bn_mul_dig(&numerator, &numerator, signers[j]+1);
        }
        // compute the inverse using little Fernat theorem
        bn_mxp_slide(&inv, &base, &r_2, &r);
        bn_mul(&acc, &acc, &inv);
        bn_mul(&acc, &acc, &numerator);
        bn_mod_barrt(&acc, &acc, &r, &u);
    }
    if (sign) bn_copy(res, &acc);
    else bn_sub(res, &r, &acc);

    // free the temp memory
    bn_free(&r);bn_free(&r_1);
    bn_free(&u);bn_free(&acc);
    bn_free(&inv);bn_free(&base);
    bn_free(&numerator);
}


// Computes the Langrange interpolation at zero LI(0) with regards to the points [signers(1)+1..signers(t+1)+1] 
// and their images [shares(1)..shares(t+1)], and stores the result in dest
// len is the polynomial degree 
void G1_lagrangeInterpolateAtZero(byte* dest, const byte* shares, const uint32_t* signers, const int len) {
    // computes Q(x) = A_0 + A_1*x + ... +  A_n*x^n  in G2
    // powers of x
    bn_st bn_lagr_coef;
    bn_new(&bn_lagr_coef);
    bn_new_size(&bn_lagr_coef, BITS_TO_BYTES(Fr_BITS));
    
    // temp variables
    ep_st mult, acc, share;
    ep_new(&mult);         
    ep_new(&acc);
    ep_new(&share);
    ep_set_infty(&acc);

    for (int i=0; i < len; i++) {
        _ep_read_bin_compact(&share, &shares[SIGNATURE_LEN*i], SIGNATURE_LEN);
        Zr_lagrangeCoefficientAtZero(&bn_lagr_coef, signers[i], signers, len);
        ep_mul_lwnaf(&mult, &share, &bn_lagr_coef);
        ep_add_projc(&acc, &acc, &mult);
    }
    // export the result
    _ep_write_bin_compact(dest, &acc, SIGNATURE_LEN);

    // free the temp memory
    ep2_free(&acc);
    ep2_free(&mult);
    ep2_free(&share);
    bn_free(&bn_lagr_coef);
    return;
}